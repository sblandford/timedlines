package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"
)

const sleepTimeMs = 100

func main() {
	timeStart := time.Now()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(130)
	}()

	if len(os.Args) < 4 || (os.Args[1] != "encode" && os.Args[1] != "decode") {
		log.Fatal("Usage: ", os.Args[0], " encode/decode <infile or -> <outfile or ->")
	}

	decode := os.Args[1] == "decode"

	inFilePath := os.Args[2]
	outFilePath := os.Args[3]

	inFile := os.Stdin
	if inFilePath != "-" {
		var err error
		inFile, err = os.Open(inFilePath)
		defer inFile.Close()
		if err != nil {
			log.Fatal("Unable to open input file : ", inFilePath)
		}
	}
	outFile := os.Stdout
	if outFilePath != "-" {
		var err error
		outFile, err = os.Create(outFilePath)
		defer outFile.Close()
		if err != nil {
			log.Fatal("Unable to open output file : ", outFilePath)
		}
	}

	reader := bufio.NewReader(inFile)
	writer := bufio.NewWriter(outFile)

	reLine := regexp.MustCompile(`^([0-9]+):([0-9]+):([0-9]+).([0-9]+):(.*)`)

	for {
		text, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		if len(text) > 0 {
			if decode {
				//Check line is in expected format
				foundLine := reLine.FindString(text)
				if len(foundLine) == 0 {
					log.Println("Line not in standard format : ", text)
					continue
				}
				// Fetch the timecode
				timePassedHoursText := reLine.ReplaceAllString(foundLine, "$1")
				timePassedMinutesText := reLine.ReplaceAllString(foundLine, "$2")
				timePassedSecondsText := reLine.ReplaceAllString(foundLine, "$3")
				timePassedMilliSecondsText := reLine.ReplaceAllString(foundLine, "$4")
				timePassedHours, err1 := strconv.ParseUint(timePassedHoursText, 10, 64)
				timePassedMinutes, err2 := strconv.ParseUint(timePassedMinutesText, 10, 64)
				timePassedSeconds, err3 := strconv.ParseUint(timePassedSecondsText, 10, 64)
				timePassedMilliSeconds, err4 := strconv.ParseUint(timePassedMilliSecondsText, 10, 64)
				if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
					log.Printf("Unable to parse time  : %s", text)
					continue
				}
				timePassedMs := timePassedMilliSeconds + (timePassedSeconds * 1000) + (timePassedMinutes * 1000 * 60) + (timePassedHours * 1000 * 3600)
				timePassed := timePassedMs * 1000000
				// Wait until time to print
				for timePassed > uint64(time.Since(timeStart)) {
					time.Sleep(sleepTimeMs * time.Millisecond)
				}

				// Fetch the line
				outText := reLine.ReplaceAllString(foundLine, "$5")

				fmt.Fprintf(writer, "%s\n", outText)
			} else {
				timePassedMs := time.Since(timeStart) / 1000000
				timePassedMilliSeconds := timePassedMs % 100
				timePassedSeconds := (timePassedMs / 1000) % 60
				timePassedMinutes := (timePassedMs / (60 * 1000)) % 60
				timePassedHours := timePassedMs / (60 * 60 * 1000)
				fmt.Fprintf(writer, "%02d:%02d:%02d.%03d:%s", timePassedHours, timePassedMinutes, timePassedSeconds, timePassedMilliSeconds, text)
			}
			writer.Flush()
		}
	}
}
