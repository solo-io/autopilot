// TODO: clean up this file
package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/solo-io/autopilot/pkg/defaults"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/rogpeppe/go-internal/modfile"
	log "github.com/sirupsen/logrus"
)

const (
	GoPathEnv  = "GOPATH"
	GoFlagsEnv = "GOFLAGS"
	GoModEnv   = "GO111MODULE"
	SrcDir     = "src"

	fsep          = string(filepath.Separator)
	mainFile      = "cmd" + fsep + "manager" + fsep + "main.go"
	rolesDir      = "roles"
	helmChartsDir = "helm-charts"
	goModFile     = "go.mod"
)

// MustInProjectRoot checks if the current dir is the project root, and exits
// if not.
func MustInProjectRoot() {
	if err := CheckProjectRoot(); err != nil {
		log.Fatal(err)
	}
}

// CheckProjectRoot checks if the current dir is the project root, and returns
// an error if not.
func CheckProjectRoot() error {
	// If the current directory has a "build/Dockerfile", then it is safe to say
	// we are at the project root.
	if _, err := os.Stat(defaults.AutopilotFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("must run command in project root dir: project structure requires %s", defaults.AutopilotFile)
		}
		return errors.Wrap(err, "error while checking if current directory is the project root")
	}
	return nil
}

func MustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: (%v)", err)
	}
	return wd
}

// gets the directory
func MustGetFileDir() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: (%v)", err)
	}
	return wd
}

// returns absolute path to the .go file containing the calling function
func MustGetThisFile() string {
	_, thisFile, _, ok := runtime.Caller(1)
	if !ok {
		log.Fatalf("Failed to get runtime.Caller")
	}
	abs, err := filepath.Abs(thisFile)
	if err != nil {
		log.Fatalf("Failed to get absolute path: (%v)", err)
	}
	return abs
}

// returns absolute path to the diretory containing the .go file containing the calling function
func MustGetThisDir() string {
	_, thisFile, _, ok := runtime.Caller(1)
	if !ok {
		log.Fatalf("Failed to get runtime.Caller")
	}
	return filepath.Dir(thisFile)
}

func getHomeDir() (string, error) {
	hd, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return homedir.Expand(hd)
}

// GetGoPkg returns the current directory's import path by parsing it from
// wd if this project's repository path is rooted under $GOPATH/src, or
// from go.mod the project uses Go modules to manage dependencies.
//
// Example: "github.com/example-inc/app-operator"
func GetGoPkg() string {
	// Default to reading from go.mod, as it should usually have the (correct)
	// package path, and no further processing need be done on it if so.
	if _, err := os.Stat(goModFile); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to read go.mod: %v", err)
	} else if err == nil {
		b, err := ioutil.ReadFile(goModFile)
		if err != nil {
			log.Fatalf("Read go.mod: %v", err)
		}
		mf, err := modfile.Parse(goModFile, b, nil)
		if err != nil {
			log.Fatalf("Parse go.mod: %v", err)
		}
		if mf.Module != nil && mf.Module.Mod.Path != "" {
			return mf.Module.Mod.Path
		}
	}

	// Then try parsing package path from $GOPATH (set env or default).
	goPath, ok := os.LookupEnv(GoPathEnv)
	if !ok || goPath == "" {
		hd, err := getHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		goPath = filepath.Join(hd, "go", "src")
	} else {
		// MustSetWdGopath is necessary here because the user has set GOPATH,
		// which could be a path list.
		goPath = MustSetWdGopath(goPath)
	}
	if !strings.HasPrefix(MustGetwd(), goPath) {
		log.Fatal("Could not determine project repository path: $GOPATH not set, wd in default $HOME/go/src, or wd does not contain a go.mod")
	}
	return parseGoPkg(goPath)
}

func parseGoPkg(gopath string) string {
	goSrc := filepath.Join(gopath, SrcDir)
	wd := MustGetwd()
	pathedPkg := strings.Replace(wd, goSrc, "", 1)
	// Make sure package only contains the "/" separator and no others, and
	// trim any leading/trailing "/".
	return strings.Trim(filepath.ToSlash(pathedPkg), "/")
}

// MustSetWdGopath sets GOPATH to the first element of the path list in
// currentGopath that prefixes the wd, then returns the set path.
// If GOPATH cannot be set, MustSetWdGopath exits.
func MustSetWdGopath(currentGopath string) string {
	var (
		newGopath   string
		cwdInGopath bool
		wd          = MustGetwd()
	)
	for _, newGopath = range filepath.SplitList(currentGopath) {
		if strings.HasPrefix(filepath.Dir(wd), newGopath) {
			cwdInGopath = true
			break
		}
	}
	if !cwdInGopath {
		log.Fatalf("Project not in $GOPATH")
	}
	if err := os.Setenv(GoPathEnv, newGopath); err != nil {
		log.Fatal(err)
	}
	return newGopath
}

var flagRe = regexp.MustCompile("(.* )?-v(.* )?")

// SetGoVerbose sets GOFLAGS="${GOFLAGS} -v" if GOFLAGS does not
// already contain "-v" to make "go" command output verbose.
func SetGoVerbose() error {
	gf, ok := os.LookupEnv(GoFlagsEnv)
	if !ok || len(gf) == 0 {
		return os.Setenv(GoFlagsEnv, "-v")
	}
	if !flagRe.MatchString(gf) {
		return os.Setenv(GoFlagsEnv, gf+" -v")
	}
	return nil
}
