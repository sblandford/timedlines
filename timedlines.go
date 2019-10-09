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

	reLine := regexp.MustCompile(`^([0-9]+)(:)(.*)`)

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
				timePassedText := reLine.ReplaceAllString(foundLine, "$1")
				timePassed, err := strconv.ParseUint(timePassedText, 10, 64)
				if err != nil {
					log.Printf("Unable to parse time (%s, %d) on line : %s", timePassedText, timePassed, text)
					continue
				}

				// Wait until time to print
				for timePassed > uint64(time.Since(timeStart)) {
					time.Sleep(sleepTimeMs * time.Millisecond)
				}

				// Fetch the line
				outText := reLine.ReplaceAllString(foundLine, "$3")

				fmt.Fprintf(writer, "%s\n", outText)
			} else {
				timePassed := time.Since(timeStart)
				fmt.Fprintf(writer, "%d:%s", timePassed, text)
			}
			writer.Flush()
		}
	}
}
