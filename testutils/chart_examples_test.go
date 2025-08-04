package testutils_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	helm "github.com/adevinta/go-helm-toolkit"
	helmtestutils "github.com/adevinta/go-helm-toolkit/testutils"
	k8stestutils "github.com/adevinta/go-k8s-toolkit/testutils"
	testutils "github.com/adevinta/go-testutils-toolkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func ExampleInstallFilteredHelmChart() {
	// use the real *testing.T from the test
	t := &testutils.FakeTest{Name: "TestICanInstallHelmChart"}

	h, err := helm.Default()
	if err != nil {
		return
	}
	fd, err := os.CreateTemp("", "")
	if err != nil {
		return
	}
	defer fd.Close()
	stdout := os.Stdout
	os.Stdout = fd
	h.RepoAdd("examples", "https://helm.github.io/examples")
	os.Stdout = stdout

	client := fake.NewClientBuilder().Build()

	name := "not-found"

	helmtestutils.InstallFilteredHelmChart(t, context.Background(), client, "test-namespace", "release-name", "examples/hello-world", k8stestutils.ExtractObjectName(k8stestutils.WithKind("Deployment"), &name))

	fmt.Println("name:", name)

	// Usually, this is done by the go framework
	fmt.Println(t)
	// Output:
	// name: release-name-hello-world
	// --- PASS: TestICanInstallHelmChart
}

// start ReadMe examples

func TestICanInstallHelmChart(t *testing.T) {

	h, err := helm.Default()
	require.NoError(t, err)
	fd, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer fd.Close()
	stdout := os.Stdout
	os.Stdout = fd
	err = h.RepoAdd("examples", "https://helm.github.io/examples")
	require.NoError(t, err)
	os.Stdout = stdout

	client := fake.NewClientBuilder().Build()

	name := "not-found"

	helmtestutils.InstallFilteredHelmChart(t, context.Background(), client, "test-namespace", "release-name", "examples/hello-world", k8stestutils.ExtractObjectName(k8stestutils.WithKind("Deployment"), &name))

	assert.Equal(t, "release-name-hello-world", name)
}
