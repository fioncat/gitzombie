package git

import (
	"fmt"
	"strings"
)

const (
	RemoteStatusSync     = "sync"
	RemoteStatusGone     = "deleted"
	RemoteStatusAhead    = "ahead"
	RemoteStatusBehind   = "behind"
	RemoteStatusConflict = "conflict"
	RemoteStatusDetached = "detached"
	RemoteStatusNone     = "none"
)

type BranchDetail struct {
	Name string

	RemoteStatus string

	Current bool

	IsRemote bool
	Remote   string

	Commit    string
	CommitMsg string
}

func ParseBranchDetail(line string) (*BranchDetail, error) {
	raw := line
	invalidLine := func(msg string) (*BranchDetail, error) {
		return nil, fmt.Errorf("invalid branch line %q, please check your git command: %s", raw, msg)
	}
	d := new(BranchDetail)
	if strings.HasPrefix(line, "*") {
		d.Current = true
		line = trimPrefix(line, "*")
	}

	if strings.HasPrefix(line, "(") {
		d.Name, line = nextRangeField(line, ")")
		d.RemoteStatus = RemoteStatusDetached
	} else {
		d.Name, line = nextField(line)
	}
	if d.Name == "" {
		return invalidLine("name is empty")
	}
	d.Commit, line = nextField(line)

	if strings.HasPrefix(line, "[") {
		var remoteDesc string
		remoteDesc, line = nextRangeField(line, "]")
		remoteDesc = trimPrefix(remoteDesc, "[")
		remoteDesc = strings.TrimSuffix(remoteDesc, "]")
		if remoteDesc == "" {
			return invalidLine("remote desc is empty")
		}

		var remoteName string
		remoteName, remoteDesc = nextField(remoteDesc)
		remoteName = strings.TrimSuffix(remoteName, ":")
		if remoteName == "" {
			return invalidLine("remote name is empty")
		}

		ahead := strings.Contains(remoteDesc, "ahead")
		behind := strings.Contains(remoteDesc, "behind")

		switch {
		case strings.Contains(remoteDesc, "gone"):
			d.RemoteStatus = RemoteStatusGone

		case ahead && behind:
			d.RemoteStatus = RemoteStatusConflict

		case ahead:
			d.RemoteStatus = RemoteStatusAhead

		case behind:
			d.RemoteStatus = RemoteStatusBehind

		default:
			d.RemoteStatus = RemoteStatusSync
		}

		d.Remote = remoteName
		d.IsRemote = true
	} else if d.RemoteStatus == "" {
		d.RemoteStatus = RemoteStatusNone
	}

	d.CommitMsg = line

	return d, nil
}

func trimPrefix(s, prefix string) string {
	s = strings.TrimPrefix(s, prefix)
	return strings.TrimSpace(s)
}

func nextRangeField(s, end string) (string, string) {
	fields := strings.Fields(s)
	var fits []string
	var fitEndIdx int
	for idx, field := range fields {
		fits = append(fits, field)
		if strings.HasSuffix(field, end) {
			fitEndIdx = idx
			break
		}
	}
	var remain string
	if fitEndIdx < len(fields)-1 {
		remain = strings.Join(fields[fitEndIdx+1:], " ")
	}
	ele := strings.Join(fits, " ")
	return ele, remain
}

func nextField(s string) (string, string) {
	fields := strings.Fields(s)
	if len(fields) == 1 {
		return s, ""
	}

	return fields[0], strings.Join(fields[1:], " ")
}

func ListLocalBranches(opts *Options) ([]*BranchDetail, error) {
	lines, err := OutputItems([]string{"branch", "-vv"}, opts)
	if err != nil {
		return nil, err
	}
	branches := make([]*BranchDetail, len(lines))
	for i, line := range lines {
		branch, err := ParseBranchDetail(line)
		if err != nil {
			return nil, err
		}
		branches[i] = branch
	}

	return branches, nil
}

func ListRemoteBranches(remote string, locals []*BranchDetail, opts *Options) ([]string, error) {
	localMap := make(map[string]struct{}, len(locals))
	for _, b := range locals {
		if b.IsRemote {
			localMap[b.Remote] = struct{}{}
		}
	}

	lines, err := OutputItems([]string{"branch", "-r"}, opts)
	if err != nil {
		return nil, err
	}

	prefix := remote + "/"
	var names []string
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			if _, ok := localMap[line]; !ok {
				name := strings.TrimPrefix(line, prefix)
				if strings.HasPrefix(name, "HEAD ->") {
					continue
				}
				names = append(names, name)
			}
		}
	}
	return names, nil
}

func ListLocalBranchNames(current bool, opts *Options) ([]string, error) {
	branches, err := ListLocalBranches(opts)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(branches))
	for _, branch := range branches {
		if branch.Current && !current {
			continue
		}
		names = append(names, branch.Name)
	}
	return names, nil
}
