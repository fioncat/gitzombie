package core

import (
	"os"
	"testing"

	"github.com/fioncat/gitzombie/pkg/osutil"
)

func TestBuilder(t *testing.T) {
	repo := &LocalRepository{
		Name:      "example/test",
		Group:     "example",
		Base:      "test",
		Path:      "/tmp/gitzombie-test/example/test",
		GroupPath: "/tmp/gitzombie-test/example",
	}
	exists, err := osutil.DirExists("/tmp/gitzombie-test")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		err = os.RemoveAll("/tmp/gitzombie-test")
		if err != nil {
			t.Fatal(err)
		}
	}
	err = repo.Setenv()
	if err != nil {
		t.Fatal(err)
	}

	err = DefaultBuilder.Run(repo)
	if err != nil {
		t.Fatal(err)
	}
}
