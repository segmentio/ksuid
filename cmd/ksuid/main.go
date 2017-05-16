package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
	"time"

	"github.com/segmentio/ksuid"
)

func main() {
	var count int
	var format string
	var template string

	flag.IntVar(&count, "n", 1, "Number of KSUIDs to generate when called with no other arguments")
	flag.StringVar(&format, "f", "inspect", "One of inspect, time, timestamp, payload, raw, or template")
	flag.StringVar(&template, "t", "", "The Go template used to format the output")
	flag.Parse()
	args := flag.Args()

	var print func(ksuid.KSUID)
	switch format {
	case "inspect":
		print = printInspect
	case "time":
		print = printTime
	case "timestamp":
		print = printTimestamp
	case "payload":
		print = printPayload
	case "raw":
		print = printRaw
	case "template":
		print = func(id ksuid.KSUID) { printTemplate(id, template) }
	default:
		fmt.Println("Bad formatting function:", format)
		os.Exit(1)
	}

	if len(args) == 0 {
		for i := 0; i < count; i++ {
			fmt.Println(ksuid.New())
		}
		os.Exit(0)
	}

	var ids []ksuid.KSUID
	for _, arg := range args {
		id, err := ksuid.Parse(arg)
		if err != nil {
			fmt.Printf("Error when parsing %q: %s\n\n", arg, err)
			flag.PrintDefaults()
			os.Exit(1)
		}
		ids = append(ids, id)
	}

	for _, id := range ids {
		print(id)
	}
}

func printInspect(id ksuid.KSUID) {
	const inspectFormat = `REPRESENTATION:

  String: %v
     Raw: %v

COMPONENTS:

       Time: %v
  Timestamp: %v
    Payload: %v

`
	fmt.Printf(inspectFormat,
		id.String(),
		strings.ToUpper(hex.EncodeToString(id.Bytes())),
		id.Time(),
		id.Timestamp(),
		strings.ToUpper(hex.EncodeToString(id.Payload())),
	)
}

func printTime(id ksuid.KSUID) {
	fmt.Println(id.Time())
}

func printTimestamp(id ksuid.KSUID) {
	fmt.Println(id.Timestamp())
}

func printPayload(id ksuid.KSUID) {
	os.Stdout.Write(id.Payload())
}

func printRaw(id ksuid.KSUID) {
	os.Stdout.Write(id.Bytes())
}

func printTemplate(id ksuid.KSUID, format string) {
	b := &bytes.Buffer{}
	t := template.Must(template.New("").Parse(format))
	t.Execute(b, struct {
		String    string
		Raw       string
		Time      time.Time
		Timestamp uint32
		Payload   string
	}{
		String:    id.String(),
		Raw:       strings.ToUpper(hex.EncodeToString(id.Bytes())),
		Time:      id.Time(),
		Timestamp: id.Timestamp(),
		Payload:   strings.ToUpper(hex.EncodeToString(id.Payload())),
	})
	b.WriteByte('\n')
	io.Copy(os.Stdout, b)
}
