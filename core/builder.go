package core

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize/english"
	"github.com/fioncat/gitzombie/config"
	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/exec"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
	"gopkg.in/yaml.v3"
)

var DefaultBuilder = func() *Builder {
	data := []byte(config.DefaultBuilder)
	var b Builder
	err := yaml.Unmarshal(data, &b)
	if err != nil {
		panic("failed to parse defaultBuilder: " + err.Error())
	}
	return &b
}()

type Builder struct {
	Create []string     `yaml:"create"`
	Init   *BuilderInit `yaml:"init"`
}

type BuilderInit struct {
	Git   *BuilderGit    `yaml:"git"`
	Steps []*BuilderStep `yaml:"steps"`

	SkipGit   bool `yaml:"skip_git"`
	SkipSteps bool `yaml:"skip_steps"`
}

type BuilderGit struct {
	Branch string `yaml:"branch"`
	User   string `yaml:"user"`
	Email  string `yaml:"email"`
	Remote string `yaml:"remote"`
	URL    string `yaml:"url"`
}

type BuilderStep struct {
	Run  string       `yaml:"run"`
	File *BuilderFile `yaml:"file"`
	From string       `yaml:"from"`
}

type BuilderFile struct {
	Name    string `yaml:"name"`
	Content string `yaml:"content"`
	From    string `yaml:"from"`
}

func GetBuilder(name string) (*Builder, error) {
	return getConfigObject[Builder](name, "builders", "builder", ".yaml", nil)
}

func ListBuilderNames() ([]string, error) {
	return listConfigObjectNames("builders", ".yaml")
}

func (b *Builder) Run(repo *LocalRepository) error {
	exists, err := osutil.DirExists(repo.Path)
	if err != nil {
		return errors.Trace(err, "check repo exists")
	}
	if exists {
		return fmt.Errorf("%s is already exists, no need to create", repo.Name)
	}

	err = osutil.EnsureDir(repo.GroupPath)
	if err != nil {
		return errors.Trace(err, "ensure group dir")
	}

	cmdWord := english.Plural(len(b.Create), "create command", "")
	term.PrintOperation("change dir to %s", repo.GroupPath)
	err = os.Chdir(repo.GroupPath)
	if err != nil {
		return errors.Trace(err, "change dir to group")
	}

	term.PrintOperation("begin to execute %s", cmdWord)
	for _, cmd := range b.Create {
		err = exec.Run(cmd, false)
		if err != nil {
			return err
		}
	}

	// After create task done, the repo dir should be created. Do a double check.
	exists, err = osutil.DirExists(repo.Path)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("path %q is still not exists after creating", repo.Path)
	}

	if !b.needRunInitGit() && !b.needRunInitSteps() {
		return nil
	}

	term.PrintOperation("change dir to %s", repo.Path)
	err = os.Chdir(repo.Path)
	if err != nil {
		return errors.Trace(err, "change dir to repo")
	}
	return b.runInit(repo)
}

func (b *Builder) needRunInitGit() bool {
	return b.Init != nil && b.Init.Git != nil && !b.Init.SkipGit
}

func (b *Builder) needRunInitSteps() bool {
	return b.Init != nil && len(b.Init.Steps) > 0 && !b.Init.SkipSteps
}

func (b *Builder) runInit(repo *LocalRepository) error {
	if b.needRunInitGit() {
		err := b.runInitGit(repo)
		if err != nil {
			return err
		}
	}

	if b.needRunInitSteps() {
		err := b.runInitSteps(repo)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Builder) runInitGit(repo *LocalRepository) error {
	term.PrintOperation("begin to init git")
	branch := b.Init.Git.Branch
	cmd := "git init"
	if branch != "" {
		cmd = fmt.Sprintf("%s -b %s", cmd, branch)
	}
	err := exec.Run(cmd, false)
	if err != nil {
		return err
	}

	user := b.Init.Git.User
	user = os.ExpandEnv(user)
	if user != "" {
		cmd = fmt.Sprintf("git config user.name %s", user)
		err = exec.Run(cmd, false)
		if err != nil {
			return err
		}
	}

	email := b.Init.Git.Email
	email = os.ExpandEnv(email)
	if email != "" {
		cmd = fmt.Sprintf("git config user.email %s", email)
		err = exec.Run(cmd, false)
		if err != nil {
			return err
		}
	}

	url := b.Init.Git.URL
	url = os.ExpandEnv(url)
	if url != "" {
		remote := b.Init.Git.Remote
		if remote == "" {
			remote = "origin"
		}
		cmd = fmt.Sprintf("git remote add %s %s", remote, email)
		err = exec.Run(cmd, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) runInitSteps(repo *LocalRepository) error {
	stepWord := english.Plural(len(b.Init.Steps), "init step", "")
	term.PrintOperation("begin to execute %s", stepWord)
	for _, step := range b.Init.Steps {
		err := b.runInitStep(repo, step)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) runInitStep(repo *LocalRepository, step *BuilderStep) error {
	if step.Run != "" {
		err := exec.Run(step.Run, false)
		if err != nil {
			return err
		}
	}
	if step.File != nil {
		err := b.runInitFile(repo, step.File)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) runInitFile(repo *LocalRepository, file *BuilderFile) error {
	file.From = os.ExpandEnv(file.From)
	file.Name = os.ExpandEnv(file.Name)
	if file.Name == "" {
		if file.From == "" {
			return errors.New("file step name and from cannot be both empty")
		}
		file.Name = filepath.Base(file.From)
	}
	var reader io.Reader
	switch {
	case file.Content != "":
		content := os.ExpandEnv(file.Content)
		var buffer bytes.Buffer
		buffer.WriteString(content)
		reader = &buffer

	case file.From != "":
		srcFile, err := os.Open(file.From)
		if err != nil {
			return err
		}
		defer srcFile.Close()
		reader = srcFile

	default:
		return errors.New("file step requires content or from")
	}

	path := filepath.Join(repo.Path, file.Name)
	dir := filepath.Dir(path)
	err := osutil.EnsureDir(dir)
	if err != nil {
		return errors.Trace(err, "file step %s: ensure dir", file.Name)
	}

	dstFile, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Trace(err, "file step %s: open file", file.Name)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, reader)
	if err != nil {
		return errors.Trace(err, "file step %s: write file", file.Name)
	}

	term.PrintCmd("write file %s", file.Name)
	return nil
}
