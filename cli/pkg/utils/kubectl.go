package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

func KubectlApply(manifest []byte, extraArgs ...string) error {
	return Kubectl(bytes.NewBuffer(manifest), append([]string{"apply", "-f", "-"}, extraArgs...)...)
}

func KubectlDelete(manifest []byte, extraArgs ...string) error {
	return Kubectl(bytes.NewBuffer(manifest), append([]string{"delete", "-f", "-"}, extraArgs...)...)
}

func Kubectl(stdin io.Reader, args ...string) error {
	kubectl := KubectlCmd(stdin, args...)

	return kubectl.Run()
}

func KubectlOut(stdin io.Reader, args ...string) (string, error) {
	kubectl := KubectlCmd(stdin, args...)

	out := &bytes.Buffer{}
	kubectl.Stdout = out
	kubectl.Stderr = out
	err := kubectl.Run()

	if err != nil {
		return "", errors.Wrap(err, out.String())
	}

	return out.String(), nil
}

func KubectlCmd(stdin io.Reader, args ...string) *exec.Cmd {
	kubectl := exec.Command("kubectl", args...)
	if stdin != nil {
		kubectl.Stdin = stdin
	}
	kubectl.Stdout = os.Stdout
	kubectl.Stderr = os.Stderr
	kubectl.Env = os.Environ()
	return kubectl
}
