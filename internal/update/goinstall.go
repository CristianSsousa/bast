package update

import (
	"fmt"
	"io"
	"os/exec"
)

// RunGoInstall executa go install module@tag, redirecionando stdout/stderr.
func RunGoInstall(tag string, stdout, stderr io.Writer) error {
	if !ValidTag(tag) {
		return fmt.Errorf("tag de versão inválida: %q", tag)
	}
	//nolint:gosec // tag validada por ValidTag
	c := exec.Command("go", "install", fmt.Sprintf("%s@%s", ModulePath, tag))
	c.Stdout = stdout
	c.Stderr = stderr
	return c.Run()
}
