package main

import (
	"os"

	"github.com/prattmic/restic-remote/binary"
	"github.com/prattmic/restic-remote/log"
)

const symbolName = "main.versionStr.str"

func main() {
	name := os.Args[1]

	v, err := binary.StringSymbol(name, symbolName)
	if err != nil {
		log.Exitf("Error reading symbol: %v", err)
	}
	log.Infof("value: %s", v)
}
