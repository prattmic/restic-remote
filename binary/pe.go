package binary

import (
	"bufio"
	"debug/pe"
	"fmt"
	"io"
)

func peStringSymbol(f *pe.File, name string) (string, error) {
	var sym *pe.Symbol
	for _, s := range f.Symbols {
		if s.Name == name {
			sym = s
			break
		}
	}

	if sym == nil {
		return "", fmt.Errorf("symbol %s not found", name)
	}

	sn := sym.SectionNumber - 1
	if sn < 0 || int(sn) >= len(f.Sections) {
		return "", fmt.Errorf("invalid section number %d for symbol %+v sections %+v", sn, sym, f.Sections)
	}
	sect := f.Sections[sn]

	sr := sect.Open()
	if _, err := sr.Seek(int64(sym.Value), io.SeekStart); err != nil {
		return "", fmt.Errorf("error seeking to %#x: %v", sym.Value, err)
	}

	r := bufio.NewReader(&io.LimitedReader{sr, maxLength})
	return r.ReadString(0)
}
