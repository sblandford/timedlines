package main

import (
	"bufio"
	"errors"
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

func timeToMs(timeText string) (int64, error) {
	reTime := regexp.MustCompile(`^(-?)([0-9]+):([0-9]+):([0-9]+).([0-9]+)$`)
	// Translate sign
	timeCodeSign := int64(1)
	if reTime.ReplaceAllString(timeText, "$1") == "-" {
		timeCodeSign = -1
	}
	// Translate timecode to mS
	timeCodeHoursText := reTime.ReplaceAllString(timeText, "$2")
	timeCodeMinutesText := reTime.ReplaceAllString(timeText, "$3")
	timeCodeSecondsText := reTime.ReplaceAllString(timeText, "$4")
	timeCodeMilliSecondsText := reTime.ReplaceAllString(timeText, "$5")
	timeCodeHours, err1 := strconv.ParseInt(timeCodeHoursText, 10, 64)
	timeCodeMinutes, err2 := strconv.ParseInt(timeCodeMinutesText, 10, 64)
	timeCodeSeconds, err3 := strconv.ParseInt(timeCodeSecondsText, 10, 64)
	timeCodeMilliSeconds, err4 := strconv.ParseInt(timeCodeMilliSecondsText, 10, 64)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return 0, errors.New("Unable to parse time as hh:mm:ss.ms format : " + timeText)
	}
	return (timeCodeMilliSeconds + (timeCodeSeconds * 1000) + (timeCodeMinutes * 1000 * 60) + (timeCodeHours * 1000 * 3600)) * timeCodeSign, nil
}

func main() {
	timeStart := time.Now()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(130)
	}()

	if len(os.Args) < 4 || (os.Args[1] != "encode" && os.Args[1] != "decode") {
		log.Fatal("Usage: ", os.Args[0], " encode/decode <infile | -> <outfile | -> [offset hh:mm:ss.ms]")
	}

	decode := os.Args[1] == "decode"

	inFilePath := os.Args[2]
	outFilePath := os.Args[3]

	offSetMs := int64(0)
	if len(os.Args) == 5 {
		var err error
		offSetMs, err = timeToMs(os.Args[4])
		if err != nil {
			log.Fatal(err)
		}
	}

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

	for {
		text, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		if len(text) > 0 {
			if decode {
				reLine := regexp.MustCompile(`^(-?)([0-9]+:[0-9]+:[0-9]+.[0-9]+):(.*)`)
				//Check line is in expected format
				foundLine := reLine.FindString(text)
				if len(foundLine) == 0 {
					log.Println("Line not in standard format : ", text)
					continue
				}
				// Fetch the timecode
				timeCodeMs, err := timeToMs(reLine.ReplaceAllString(foundLine, "$2"))
				if err != nil {
					log.Println(err)
					continue
				}
				timeCode := (timeCodeMs - offSetMs) * 1000000
				// Wait until time to print
				for timeCode > int64(time.Since(timeStart)) {
					time.Sleep(sleepTimeMs * time.Millisecond)
				}

				// Fetch the line
				outText := reLine.ReplaceAllString(foundLine, "$3")

				fmt.Fprintf(writer, "%s\n", outText)
			} else {
				timeCodeMs := int64((time.Since(timeStart) / 1000000)) + offSetMs
				if timeCodeMs < 0 {
					timeCodeMs = -timeCodeMs
					fmt.Fprintf(writer, "-")
				}
				timeCodeMilliSeconds := timeCodeMs % 100
				timeCodeSeconds := (timeCodeMs / 1000) % 60
				timeCodeMinutes := (timeCodeMs / (60 * 1000)) % 60
				timeCodeHours := timeCodeMs / (60 * 60 * 1000)
				fmt.Fprintf(writer, "%02d:%02d:%02d.%03d:%s", timeCodeHours, timeCodeMinutes, timeCodeSeconds, timeCodeMilliSeconds, text)
			}
			writer.Flush()
		}
	}
}
