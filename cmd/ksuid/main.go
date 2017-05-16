package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/segmentio/ksuid"
)

const usage = `
Usage: ksuid <ksuid>
       ksuid
			 ksuid --help

Generates and inspects KSUIDs
`

const inspectFormat = `
REPRESENTATION:

  String: %v
     Raw: %v

COMPONENTS:

       Time: %v
  Timestamp: %v
    Payload: %v

`

func main() {
	if len(os.Args) < 2 {
		fmt.Println(ksuid.New())
		os.Exit(0)
	}

	id, err := ksuid.Parse(os.Args[1])
	if err != nil {
		fmt.Println("Error when parsing: ", err)
		fmt.Println()
		fmt.Println(usage)
		os.Exit(1)
	}

	fmt.Printf(inspectFormat,
		id.String(),
		strings.ToUpper(hex.EncodeToString(id.Bytes())),
		id.Time(),
		id.Timestamp(),
		strings.ToUpper(hex.EncodeToString(id.Payload())))

}
