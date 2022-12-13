package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
)

var templateNameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func templatePath(name string, ensureExists bool) (string, error) {
	if name == "" {
		return "", errors.New("template name cannot be empty")
	}
	if !templateNameRe.MatchString(name) {
		return "", fmt.Errorf("invalid template name %q", name)
	}

	baseDir := config.GetDir("templates")
	err := osutil.EnsureDir(baseDir)
	if err != nil {
		return "", errors.Trace(err, "ensure template dir")
	}

	path := filepath.Join(baseDir, name)
	exists, err := osutil.DirExists(path)
	if err != nil {
		return "", errors.Trace(err, "check template exists")
	}
	if !exists && ensureExists {
		return "", fmt.Errorf("cannot find template %q", name)
	}
	if exists && !ensureExists {
		return "", fmt.Errorf("template %q is already exists", name)
	}
	return path, nil
}

func GetTemplate(name string) (string, error) {
	return templatePath(name, true)
}

type TemplateFile struct {
	Name string

	Read io.ReadCloser
	Mode os.FileMode
}

func (file *TemplateFile) WriteTo(dstRoot string) error {
	path := filepath.Join(dstRoot, file.Name)
	dirPath := filepath.Dir(path)
	err := osutil.EnsureDir(dirPath)
	if err != nil {
		return errors.Trace(err, "ensure file dir")
	}

	dst, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, file.Mode)
	if err != nil {
		return err
	}
	defer dst.Close()
	defer file.Read.Close()

	_, err = io.Copy(dst, file.Read)
	if err != nil {
		return errors.Trace(err, "copy file %q", file.Name)
	}
	return nil
}

func FindTemplateFiles(root string) ([]*TemplateFile, error) {
	es, err := os.ReadDir(root)
	if err != nil {
		return nil, errors.Trace(err, "read template root")
	}
	if len(es) == 0 {
		return nil, nil
	}

	stack := make([]string, 0, len(es)-1)
	for _, e := range es {
		name := e.Name()
		if e.IsDir() && name == ".git" {
			continue
		}
		stack = append(stack, filepath.Join(root, e.Name()))
	}

	var files []*TemplateFile
	for len(stack) > 0 {
		// pop a path, the path here MUST be abs.
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if !filepath.IsAbs(cur) {
			panic("internal: FindTemplateFiles path must be abs")
		}

		stat, err := os.Stat(cur)
		if err != nil {
			return nil, errors.Trace(err, "stat template file")
		}
		if stat.IsDir() {
			es, err = os.ReadDir(cur)
			if err != nil {
				return nil, errors.Trace(err, "read template dir")
			}
			for _, e := range es {
				stack = append(stack, filepath.Join(cur, e.Name()))
			}
			continue
		}

		name, err := filepath.Rel(root, cur)
		if err != nil {
			return nil, errors.Trace(err, "get rel template name")
		}
		name = strings.Trim(name, "/")

		src, err := os.Open(cur)
		if err != nil {
			return nil, errors.Trace(err, "open src file")
		}

		file := &TemplateFile{
			Name: name,

			Read: src,
			Mode: stat.Mode(),
		}
		files = append(files, file)
	}
	return files, nil
}

func CreateTemplate(name string, files []*TemplateFile) error {
	root, err := templatePath(name, false)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = file.WriteTo(root)
		if err != nil {
			return err
		}
		term.PrintOperation("Copy file %q", file.Name)
	}

	return nil
}

func DeleteTemplate(name string) error {
	path, err := templatePath(name, true)
	if err != nil {
		return err
	}
	term.ConfirmExit("Do you want to delete template %q", name)
	return errors.Trace(os.RemoveAll(path), "remove dir")
}

func UseTemplate(name, dst string) error {
	path, err := templatePath(name, true)
	if err != nil {
		return err
	}

	files, err := FindTemplateFiles(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = file.WriteTo(dst)
		if err != nil {
			return err
		}
		term.PrintOperation("Write file %q", file.Name)
	}

	return nil
}

func ListTemplates() ([]string, error) {
	baseDir := config.GetDir("templates")
	es, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(es))
	for _, e := range es {
		if !e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	return names, nil
}
