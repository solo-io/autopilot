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
	kubectl := exec.Command("kubectl", args...)
	if stdin != nil {
		kubectl.Stdin = stdin
	}
	kubectl.Stdout = os.Stdout
	kubectl.Stderr = os.Stderr
	kubectl.Env = os.Environ()
	return kubectl.Run()
}

func KubectlOut(stdin io.Reader, args ...string) (string, error) {
	kubectl := exec.Command("kubectl", args...)
	if stdin != nil {
		kubectl.Stdin = stdin
	}

	out := &bytes.Buffer{}

	kubectl.Stdout = out
	kubectl.Stderr = out
	kubectl.Env = os.Environ()
	err := kubectl.Run()

	if err != nil {
		return "", errors.Wrap(err, out.String())
	}

	return out.String(), nil
}
