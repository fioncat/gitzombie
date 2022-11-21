package worker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/git"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
)

type ErrorHandler struct {
	Name string

	LogPath string

	Header  func(idx int, err error) string
	Content func(idx int, err error) string
}

func HandleErrors(errs []error, h *ErrorHandler) error {
	var sb strings.Builder
	for idx, err := range errs {
		header := h.Header(idx, err)
		content := h.Content(idx, err)
		header = fmt.Sprintf("==================== %s ====================", header)
		sb.WriteString(header + "\n")
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}
	logPath := h.LogPath
	if logPath == "" {
		logPath = filepath.Join(os.TempDir(), "gitzombie", "logs", h.Name)
	}
	err := osutil.WriteFile(logPath, []byte(sb.String()))
	if err != nil {
		return errors.Trace(err, "write log file")
	}
	errWord := english.Plural(len(errs), "error", "")
	term.Print("write red|%s log| to green|%s|", errWord, logPath)
	return fmt.Errorf("workflow failed with %s", errWord)
}

func GitHeader(idx int, err error) string {
	if gitErr, ok := err.(*git.ExecError); ok {
		return fmt.Sprintf("command %q failed: %v", gitErr.Cmd, gitErr.Err)
	}
	return fmt.Sprintf("%d git command failed: %v", idx, err)
}

func GitContent(_ int, err error) string {
	if gitErr, ok := err.(*git.ExecError); ok {
		return gitErr.Stderr
	}
	return ""
}
