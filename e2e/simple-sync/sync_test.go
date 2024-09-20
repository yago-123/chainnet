//go:build e2e
// +build e2e

package simple_sync

import (
	"os"
	"testing"

	"github.com/yago-123/minikube-testing/pkg/orchestrator"
)

func TestNodeSyncDuringStartup(t *testing.T) {
	minikube := orchestrator.NewMinikube(os.Stdout, os.Stderr)
	defer minikube.Delete()

	_, err := minikube.Create("v1.31.0", 1, 5, 5120)
	if err != nil {
		t.Errorf("unable to create minikube cluster: %s", err)
	}

	client.DeployWithHelm()
}
