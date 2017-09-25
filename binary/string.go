package binary

import (
	"debug/elf"
	"debug/pe"
	"fmt"
)

// maxLength is the largest symbol size to read.
const maxLength = 1024

// StringSymbol returns the value of a NUL-terminated string symbol from an ELF
// or PE binary.
func StringSymbol(filename, sym string) (string, error) {
	if f, err := elf.Open(filename); err == nil {
		defer f.Close()
		return elfStringSymbol(f, sym)
	}
	if f, err := pe.Open(filename); err == nil {
		defer f.Close()
		return peStringSymbol(f, sym)
	}
	return "", fmt.Errorf("cannot open %s as ELF or PE", filename)
}
