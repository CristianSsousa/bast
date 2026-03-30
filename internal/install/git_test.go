package install

import (
	"testing"
)

func TestIsCommandAvailable_go(t *testing.T) {
	t.Parallel()
	if !isCommandAvailable("go") {
		t.Skip("go não está no PATH")
	}
}

func TestGetGitVersion_whenGitInstalled(t *testing.T) {
	t.Parallel()
	if !isCommandAvailable("git") {
		t.Skip("git não está no PATH")
	}
	v, err := getGitVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v == "" {
		t.Fatal("versão vazia")
	}
}
