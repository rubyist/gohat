package main

import (
	"fmt"
	"github.com/rubyist/gohat/pkg/heapfile"
)

func hexDump(content string) string {
	output := "0000000 "
	lastIdx := 0
	contentBytes := []byte(content)
	for idx, c := range contentBytes {
		output += fmt.Sprintf("%0.2x ", c)
		if (idx+1)%16 == 0 {
			for _, j := range contentBytes[lastIdx : idx+1] {
				if int(j) >= 0x20 && int(j) <= 0x7e {
					output += string(j)
				} else {
					output += "."
				}
			}
			lastIdx = idx + 1
			output += fmt.Sprintf("\n%0.7x ", idx+1)
		}
	}
	wholeRows := len(contentBytes) / 16
	lastRow := len(contentBytes) - wholeRows*16
	filler := 16 - lastRow
	for i := 0; i < filler; i++ {
		output += "   "
	}

	output += " " + string(contentBytes[lastIdx:])
	return output
}

func displayObjectShort(o *heapfile.Object) {
	typeName := "unknown"
	if o.Type != nil {
		typeName = o.Type.Name
	}
	fmt.Printf("%x %s\n", o.Address, typeName)
}
