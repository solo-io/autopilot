package main

import (
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/kubeutils"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {


}

func initMetrics() error {

}

func initMananger(kcEnv, namespace string) (manager.Manager, error) {
	kubeconfig := os.Getenv(kcEnv)
	if kubeconfig == "" {
		return nil, errors.Errorf("expected " + kcEnv + " to be a path to a kubeconfig file")
	}

	cfg, err := kubeutils.GetConfig("", kubeconfig)
	if err != nil {
		return nil, err
	}

	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		return nil, err
	}
	return mgr, nil
}
