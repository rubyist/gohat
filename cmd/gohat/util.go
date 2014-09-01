package main

import (
	"fmt"
)

func hexDump(content string) string {
	output := "0000000 "
	for idx, c := range []byte(content) {
		output += fmt.Sprintf("%.2x ", c)
		if (idx+1)%16 == 0 {
			output += fmt.Sprintf("\n%.7x ", idx+1)
		}
	}
	return output
}
