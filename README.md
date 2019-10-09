# timedlines
Records/plays back lines (e.g. from stdin) with timing information

The timing is taken at the end of each line.

## Usage

```
timelines encode - test.txt
```
Takes input from keyboard and writes timed output to test.txt

```
timelines decode test.txt -
```
Decodes the test.txt file and outputs text lines at the same pace they were created.


```
timelines encode somefile.fifo test.txt
```
Takes input from a fifo and writes timeed output to test.txt. It only makes sense to use this utility to encode with a fifo or stdin as input.


```
tail -n 1 -f /var/log/syslog | ./timedlines encode - syslog.timed.txt
```
Tails syslog and encodes events that can be played back at the same pace as they were created.

## Compiling
go build -o timedlines timedlines.go
