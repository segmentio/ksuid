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

var (
	count   int
	format  string
	tpltxt  string
	verbose bool
)

func init() {
	flag.IntVar(&count, "n", 1, "Number of KSUIDs to generate when called with no other arguments.")
	flag.StringVar(&format, "f", "string", "One of string, inspect, time, timestamp, payload, raw, or template.")
	flag.StringVar(&tpltxt, "t", "", "The Go template used to format the output.")
	flag.BoolVar(&verbose, "v", false, "Turn on verbose mode.")
}

func main() {
	flag.Parse()
	args := flag.Args()

	var print func(ksuid.KSUID)
	switch format {
	case "string":
		print = printString
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
		print = printTemplate
	default:
		fmt.Println("Bad formatting function:", format)
		os.Exit(1)
	}

	if len(args) == 0 {
		for i := 0; i < count; i++ {
			args = append(args, ksuid.New().String())
		}
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
		if verbose {
			fmt.Printf("%s: ", id)
		}
		print(id)
	}
}

func printString(id ksuid.KSUID) {
	fmt.Println(id.String())
}

func printInspect(id ksuid.KSUID) {
	const inspectFormat = `
REPRESENTATION:

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

func printTemplate(id ksuid.KSUID) {
	b := &bytes.Buffer{}
	t := template.Must(template.New("").Parse(tpltxt))
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
