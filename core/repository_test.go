package core

import (
	"reflect"
	"testing"

	"github.com/fioncat/gitzombie/config"
)

func TestRepositoryStorageAdd(t *testing.T) {
	err := config.Init()
	if err != nil {
		t.Fatal(err)
	}
	remote := &Remote{
		Name: "github",
		Host: "github.com",

		Protocol: "https",

		User:  "fioncat",
		Email: "lazycat7706@gmail.com",

		Provider: "github",
	}

	repoNames := []string{
		"fioncat/gitzombie",
		"fioncat/dotfiles",
		"kubernetes/kubernetes",
		"spf13/cobra",
	}
	s, err := NewRepositoryStorage()
	if err != nil {
		t.Fatal(err)
	}

	expects := make([]*Repository, len(repoNames))
	for i, repoName := range repoNames {
		repo, err := CreateRepository(remote, repoName)
		if err != nil {
			t.Fatal(err)
		}
		expects[i] = repo
		if _, err = s.GetByName(remote.Name, repoName); err == nil {
			continue
		}
		err = s.Add(repo)
		if err != nil {
			t.Fatal(err)
		}
	}
	err = s.Close()
	if err != nil {
		t.Fatal(err)
	}

	s, err = NewRepositoryStorage()
	if err != nil {
		t.Fatal(err)
	}

	for i, repoName := range repoNames {
		repo, err := s.GetByName(remote.Name, repoName)
		if err != nil {
			t.Fatal(err)
		}
		expect := expects[i]
		if !reflect.DeepEqual(expect, repo) {
			t.Fatalf("unexpect repo %+v", repo)
		}
	}
}
