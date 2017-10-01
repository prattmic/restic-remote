package binver

import (
	"os/exec"
	"strings"
)

func Client(bin string) (string, error) {
	cmd := exec.Command(bin, "--version")
	b, err := cmd.Output()
	return strings.Trim(string(b), "\r\n"), err
}

func Restic(bin string) (string, error) {
	cmd := exec.Command(bin, "version")
	b, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Only look at the first line. The second line contains the system the
	// binary is running on, which isn't relevant.
	s := strings.Trim(string(b), "\r\n")
	s = strings.SplitN(s, "\n", 2)[0]
	return strings.Trim(s, "\r\n"), nil
}

