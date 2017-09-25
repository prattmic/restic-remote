package binary

import (
	"bufio"
	"debug/elf"
	"fmt"
	"io"
)

func elfStringSymbol(f *elf.File, name string) (string, error) {
	syms, err := f.Symbols()
	if err != nil {
		return "", fmt.Errorf("unable to read symbol table from %+v: %v", f, err)
	}

	var sym *elf.Symbol
	for _, s := range syms {
		if s.Name == name {
			v := s
			sym = &v
			break
		}
	}

	if sym == nil {
		return "", fmt.Errorf("symbol %s not found", name)
	}

	sn := sym.Section
	if sn < 0 || int(sn) >= len(f.Sections) {
		return "", fmt.Errorf("invalid section number %d for symbol %+v sections %+v", sn, sym, f.Sections)
	}
	sect := f.Sections[sn]

	off := int64(sym.Value - sect.Addr)
	if off < 0 || uint64(off) >= sect.Size {
		return "", fmt.Errorf("symbol %+v not in section %+v", sym, sect)
	}

	sr := sect.Open()
	if _, err := sr.Seek(off, io.SeekStart); err != nil {
		return "", fmt.Errorf("error seeking to %#x: %v", sym.Value, err)
	}

	r := bufio.NewReader(&io.LimitedReader{sr, maxLength})
	return r.ReadString(0)
}
