package term

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const FilePlaceholder = "{file}"

func FuzzySearch(name string, items []string) (int, error) {
	PrintOperation("use fzf to search %s", name)
	var inputBuf bytes.Buffer
	inputBuf.Grow(len(items))
	for _, item := range items {
		inputBuf.WriteString(item + "\n")
	}

	var outputBuf bytes.Buffer
	cmd := exec.Command("fzf")
	cmd.Stdin = &inputBuf
	cmd.Stderr = os.Stderr
	cmd.Stdout = &outputBuf

	err := cmd.Run()
	if err != nil {
		return 0, err
	}

	result := outputBuf.String()
	result = strings.TrimSpace(result)
	for idx, item := range items {
		if item == result {
			return idx, nil
		}
	}

	return 0, fmt.Errorf("cannot find %q", result)
}
