package core

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/fioncat/gitzombie/pkg/errors"
	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/fioncat/gitzombie/pkg/term"
)

var MuteJob bool

type Job struct {
	Name string `yaml:"name" validate:"required"`
	Run  string `yaml:"run"`

	RequireEnv []string `yaml:"require_env"`
}

type JobError struct {
	Name string
	Path string
	Err  error
	out  string
}

func (err *JobError) Error() string {
	return fmt.Sprintf("failed to execute job %s on %s: %v", err.Name, err.Path, err.Err)
}

func (err *JobError) Out() string {
	return err.out
}

func ListJobNames() ([]string, error) {
	return listConfigObjects("jobs", shExt)
}

func GetJobPath(name string) (string, error) {
	path := getConfigObjectPath("jobs", shExt, name)
	exists, err := osutil.FileExists(path)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("cannot find job %s", name)
	}
	return path, nil
}

func (job *Job) Execute(root string, env osutil.Env) error {
	if key, skip := job.Skip(env); skip {
		job.Say("skip job %q because of unbound env %q", job.Name, key)
		return nil
	}

	var out bytes.Buffer
	cmd, err := job.Cmd(root, env, &out)
	if err != nil {
		return err
	}
	job.Say("running %s", job.Name)

	err = cmd.Run()
	if err != nil {
		return &JobError{
			Name: job.Name,
			Path: root,
			Err:  err,
			out:  out.String(),
		}
	}
	return nil
}

func (job *Job) Skip(env osutil.Env) (string, bool) {
	for _, requireKey := range job.RequireEnv {
		var ok bool
		if env != nil {
			_, ok = env[requireKey]
		}
		if !ok {
			return requireKey, true
		}
	}
	return "", false
}

func (job *Job) Say(msg string, args ...any) {
	if MuteJob {
		return
	}
	msg = fmt.Sprintf(msg, args...)
	term.PrintOperation(msg)
}

func (job *Job) Cmd(root string, env osutil.Env, out *bytes.Buffer) (*exec.Cmd, error) {
	var args []string
	if job.Run != "" {
		args = []string{"-c", job.Run}
	} else {
		jobPath, err := GetJobPath(job.Name)
		if err != nil {
			return nil, err
		}
		args = []string{jobPath}
	}

	cmd := exec.Command("bash", args...)
	if MuteJob {
		if out != nil {
			cmd.Stdout = out
			cmd.Stderr = out
		}
	} else {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
	}
	if root != "" {
		exists, err := osutil.DirExists(root)
		if err != nil {
			return nil, errors.Trace(err, "check job root path exists")
		}
		if !exists {
			return nil, fmt.Errorf("job root path %s does not exists", root)
		}
		cmd.Dir = root
	}
	if len(env) > 0 {
		env.SetCmd(cmd)
	}
	return cmd, nil
}

func wrapJobCmdError(err error) error {
	if _, ok := err.(*JobError); ok {
		return errors.New("failed to execute job")
	}
	return fmt.Errorf("failed to execute job: %v", err)
}
