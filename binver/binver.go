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
	return strings.Trim(string(b), "\r\n"), err
}

