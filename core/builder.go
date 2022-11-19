package core

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
	"github.com/fioncat/gitzombie/pkg/validate"
	"gopkg.in/yaml.v3"
)

var DefaultBuilder = func() *Builder {
	var b Builder
	err := yaml.Unmarshal([]byte(config.DefaultBuilder), &b)
	if err != nil {
		panic(err)
	}
	err = b.Validate()
	if err != nil {
		panic(err)
	}
	return &b
}()

type Builder struct {
	Create []*Job       `yaml:"create" validate:"dive"`
	Files  []*BuildFile `yaml:"files" validate:"unique=Name,dive"`
	Init   []*Job       `yaml:"init" validate:"dive"`

	repo *Repository
	env  osutil.Env
}

type BuildFile struct {
	Name    string `yaml:"name" validate:"required"`
	Content string `yaml:"content"`
	From    string `yaml:"from"`
}

func ListBuilderNames() ([]string, error) {
	return listConfigObjects("builders", yamlExt)
}

func GetBuilder(name string) (*Builder, error) {
	return getConfigObject("builders", yamlExt, "builder", name, func(b *Builder) error {
		if len(b.Create) == 0 {
			b.Create = DefaultBuilder.Create
		}
		return b.Validate()
	})
}

func (b *Builder) Validate() error {
	return validate.Do(b)
}

func (b *Builder) Prepare(remote *Remote, repo *Repository) error {
	env := make(osutil.Env)
	err := repo.SetEnv(remote, env)
	if err != nil {
		return err
	}
	b.env = env
	b.repo = repo

	for _, file := range b.Files {
		file.Name = env.Expand(file.Name)
		file.Content = env.Expand(file.Content)
		file.From = env.Expand(file.From)
	}
	return nil
}

func (b *Builder) Execute() error {
	exists, err := osutil.DirExists(b.repo.Path)
	if err != nil {
		return errors.Trace(err, "check repo exists")
	}
	if exists {
		return fmt.Errorf("repo %s is already exists, no need to create", b.repo.Name)
	}
	dir, err := b.repo.EnsureDir()
	if err != nil {
		return errors.Trace(err, "ensure repo dir")
	}
	term.PrintOperation("begin to create repo green|%s|", b.repo.Name)
	err = b.executeJobs(dir, b.Create)
	if err != nil {
		return err
	}

	exists, err = osutil.DirExists(b.repo.Path)
	if err != nil {
		return errors.Trace(err, "check repo exists")
	}
	if !exists {
		return errors.New("repo is still not exists after creating, please check your builder")
	}

	for _, file := range b.Files {
		term.PrintOperation("write file green|%s|", file.Name)
		err = b.executeFile(file)
		if err != nil {
			return err
		}
	}
	if len(b.Init) > 0 {
		term.PrintOperation("begin to init repo green|%s|", b.repo.Name)
		return b.executeJobs(b.repo.Path, b.Init)
	}

	return nil
}

func (b *Builder) executeJobs(root string, jobs []*Job) error {
	for _, job := range jobs {
		err := job.Execute(root, b.env)
		if err != nil {
			return wrapJobCmdError(err)
		}
	}
	return nil
}

func (b *Builder) executeFile(file *BuildFile) error {
	var src io.Reader
	if file.From != "" {
		srcFile, err := os.Open(file.From)
		if err != nil {
			return errors.Trace(err, "open from file")
		}
		defer srcFile.Close()
		src = srcFile
	} else {
		var buffer bytes.Buffer
		buffer.WriteString(file.Content)
		src = &buffer
	}
	path := filepath.Join(b.repo.Path, file.Name)
	dir := filepath.Dir(path)
	err := osutil.EnsureDir(dir)
	if err != nil {
		return errors.Trace(err, "ensure file dir")
	}

	dst, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return errors.Trace(err, "write file")
}
