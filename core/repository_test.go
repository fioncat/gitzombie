package core

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fioncat/gitzombie/pkg/osutil"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func writeRepoBinary(file *os.File, repos []*Repository) error {
	encoder := gob.NewEncoder(file)
	return encoder.Encode(repos)
}

func writeRepoJson(file *os.File, repos []*Repository) error {
	encoder := json.NewEncoder(file)
	return encoder.Encode(repos)
}

func writeRepoYaml(file *os.File, repos []*Repository) error {
	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	return encoder.Encode(repos)
}

func writeRepoToml(file *os.File, repos []*Repository) error {
	encoder := toml.NewEncoder(file)
	return encoder.Encode(map[string][]*Repository{"repos": repos})
}

func readRepoBinary(file *os.File) error {
	decoder := gob.NewDecoder(file)
	var repos []*Repository
	return decoder.Decode(&repos)
}

func readRepoJson(file *os.File) error {
	decoder := json.NewDecoder(file)
	var repos []*Repository
	return decoder.Decode(&repos)
}

func readRepoYaml(file *os.File) error {
	decoder := yaml.NewDecoder(file)
	var repos []*Repository
	return decoder.Decode(&repos)
}

func readRepoToml(file *os.File) error {
	var data map[string][]*Repository
	decoder := toml.NewDecoder(file)
	return decoder.Decode(&data)
}

func buildTestRepos(n int) []*Repository {
	repos := make([]*Repository, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("test/%d", i)
		repos[i] = &Repository{
			Name:   name,
			Path:   filepath.Join("/path/to/repo", name),
			Remote: "test",
			Access: uint64(i),
		}
	}
	return repos
}

var (
	testRepoDir = filepath.Join(os.TempDir(), "gitzombie-bench")

	testRepoBinaryPath = filepath.Join(testRepoDir, "binary")
	testRepoYamlPath   = filepath.Join(testRepoDir, "test.yaml")
	testRepoTomlPath   = filepath.Join(testRepoDir, "test.toml")
	testRepoJsonPath   = filepath.Join(testRepoDir, "test.json")
)

func TestMain(m *testing.M) {
	err := osutil.EnsureDir(testRepoDir)
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func BenchmarkBinary(b *testing.B) {
	repos := buildTestRepos(b.N)
	file, err := os.OpenFile(testRepoBinaryPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	err = writeRepoBinary(file, repos)
	if err != nil {
		b.Fatal(err)
	}

	file, err = os.Open(testRepoBinaryPath)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	err = readRepoBinary(file)
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkWriteJson(b *testing.B) {
	repos := buildTestRepos(b.N)
	file, err := os.OpenFile(testRepoJsonPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	err = writeRepoJson(file, repos)
	if err != nil {
		b.Fatal(err)
	}

	file, err = os.Open(testRepoJsonPath)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	err = readRepoJson(file)
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkWriteYaml(b *testing.B) {
	repos := buildTestRepos(b.N)
	file, err := os.OpenFile(testRepoYamlPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	err = writeRepoYaml(file, repos)
	if err != nil {
		b.Fatal(err)
	}

	file, err = os.Open(testRepoYamlPath)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	err = readRepoYaml(file)
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkWriteToml(b *testing.B) {
	repos := buildTestRepos(b.N)
	file, err := os.OpenFile(testRepoTomlPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	err = writeRepoToml(file, repos)
	if err != nil {
		b.Fatal(err)
	}

	file, err = os.Open(testRepoTomlPath)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	err = readRepoToml(file)
	if err != nil {
		b.Fatal(err)
	}
}
