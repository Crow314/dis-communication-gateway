package main

import (
	"flag"
	"fmt"
	"github.com/Crow314/dis-communication-repeater/pkg/repeater"
	"github.com/Crow314/im920s-controller/pkg/connector"
	"github.com/Crow314/im920s-controller/pkg/module"
	"os"
)

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, `
Usage of %s:
  %s [options] tty
Args
  tty
        TTY device path
        e.g.) /dev/ttyUSB0
Options
`, os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}

	storeSize := flag.Int("s", 256, "remember packet identifications specified number of times")
	times := flag.Int("n", 3, "resend received packet specified number of times")
	interval := flag.Int("i", 10_000, "wait specified amount of time before sending same packet")

	flag.Parse()
	tty := flag.Args()[0]

	conn := connector.NewConnector(tty)
	im920s := module.NewIm920s(conn.TransmitChannel(), conn.ReceiveChannel())

	repeater.Run(im920s, *storeSize, *times, *interval, nil)
}
