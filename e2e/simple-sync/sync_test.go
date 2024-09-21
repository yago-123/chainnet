//go:build e2e
// +build e2e

package simple_sync

import (
	"context"
	"github.com/yago-123/minikube-testing/pkg/orchestrator"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yago-123/minikube-testing/pkg/runtime"
)

const (
	ctxTimeout = 5 * time.Minute
)

func TestNodeSyncDuringStartup(t *testing.T) {
	// set working directory to root of the project
	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		panic(err)
	}

	// create minikube machine
	minikube := orchestrator.NewMinikube(os.Stdout, os.Stderr)
	defer minikube.Delete()

	_, err := minikube.Create("v1.28.3", 1, 5, 5120)
	if err != nil {
		t.Errorf("unable to create minikube cluster: %s", err)
	}

	// build docker images with current code
	imageTag := uuid.NewString()
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	dock, err := runtime.NewDockerController()
	if err != nil {
		t.Fatalf("unable to create docker controller: %v", err)
	}

	dockerfile, err := os.ReadFile("./build/docker/miner/Dockerfile")
	if err != nil {
		t.Fatalf("unable to read dockerfile: %v", err)
	}
	if err = dock.BuildImage(ctx, "yagoninja/chainnet-miner", imageTag, dockerfile, []string{"./bin/chainnet-miner"}); err != nil {
		t.Fatalf("unable to build image: %v", err)
	}

	// load docker images in minikube
	if err = minikube.LoadImage("yagoninja/chainnet-miner", imageTag); err != nil {
		t.Fatalf("unable to load image: %v", err)
	}

	// deploy pods

	// run checks

}
