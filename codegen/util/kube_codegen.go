package util

import (
	"os"
	"os/exec"
	"path/filepath"
)

func KubeCodegen(group, version, apiDir string) error {
	// TODO: replace these funcs with anyvendor

	// path on disk
	modulePath := GetModuleRoot()
	// go module package
	modulePkg := GetModulePkg()

	// input dir
	apiPkg := filepath.Join(modulePkg, apiDir)

	// output dir
	clientPkg := filepath.Join(apiPkg, group, version)

	cmd := exec.Command("bash", "-c", codegenScript)

	cmd.Env = append(os.Environ(),
		"MODULE_PATH="+modulePath,
		"MODULE_PKG="+modulePkg,
		"API_PKG="+apiPkg,
		"CLIENT_PKG="+clientPkg,
		"GROUP="+group,
		"VERSION="+version,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

var codegenScript = `
#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


# Below code is copied from https://github.com/weaveworks/flagger/blob/master/hack/update-codegen.sh
CODEGEN_PKG=$(go list -f '{{ .Dir }}' -m k8s.io/code-generator)


echo ">> Using ${CODEGEN_PKG}"

# code-generator does work with go.mod but makes assumptions about
# the project living in $GOPATH/src. To work around this and support
# any location; create a temporary directory, use this as an output
# base, and copy everything back once generated.
TEMP_DIR=$(mktemp -d)
cleanup() {
    echo ">> Removing ${TEMP_DIR}"
    rm -rf ${TEMP_DIR}
}
trap "cleanup" EXIT SIGINT

echo ">> Temporary output directory ${TEMP_DIR}"

# Ensure we can execute.
chmod +x ${CODEGEN_PKG}/generate-groups.sh


${CODEGEN_PKG}/generate-groups.sh all \
    ${CLIENT_PKG} \
    ${API_PKG} \
    ${GROUP}:${VERSION} \
    --output-base "${TEMP_DIR}"
# Copy everything back.
cp -a "${TEMP_DIR}/${MODULE_PKG}/." "${MODULE_PATH}/"
`
