package helm_test

import (
	"io/ioutil"
	"os"
	"testing"

	helm "github.com/adevinta/go-helm-toolkit"
	system "github.com/adevinta/go-system-toolkit"
	testutils "github.com/adevinta/go-testutils-toolkit"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownload(t *testing.T) {
	testID := uuid.New().String()
	defer os.RemoveAll(testID)
	path := testID + "/.helm/bin/helm-v3.3.4"
	assert.NoError(t, helm.Download("v3.3.4", path))
	h := helm.Helm3{Path: path}
	assert.Equal(t, "v3.3.4", h.Version())
}

func TestSupported(t *testing.T) {
	const renderedChart = `---
# Source: test-chart/templates/test-namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: release
  labels:
    name: release
`
	foundVersion := 0
	for h := range helm.Supported() {
		foundVersion++
		reader, err := h.Template("namespace", "release", "./test-data/chart/test-chart")
		assert.NoError(t, err)
		b, err := ioutil.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, renderedChart, string(b))
		assert.NotEqual(t, "", h.Version())
	}
	assert.Greater(t, foundVersion, 0)
}

func TestDiscoverChartDirFindsAllCharts(t *testing.T) {
	t.Cleanup(system.Reset)
	fs := afero.NewMemMapFs()
	system.DefaultFileSystem = fs
	_, err := fs.Create("/some/directory//subfolder/some/chart/Chart.yaml")
	require.NoError(t, err)
	_, err = fs.Create("/some/directory/other/chart/Chart.yaml")
	require.NoError(t, err)
	_, err = fs.Create("/some/directory/some-other-subfolder/some-folder")
	require.NoError(t, err)
	chartDirs := []string{}
	for d := range helm.DiscoverChartDirs("/some/directory") {
		chartDirs = append(chartDirs, d)
	}
	assert.Len(t, chartDirs, 2)
	assert.Contains(t, chartDirs, "/some/directory/other/chart")
	assert.Contains(t, chartDirs, "/some/directory/subfolder/some/chart")
}

func TestDiscoverChartTestsFindsAllTests(t *testing.T) {
	t.Cleanup(system.Reset)
	fs := afero.NewMemMapFs()
	system.DefaultFileSystem = fs
	_, err := fs.Create("/some/directory//subfolder/some/chart/Chart.yaml")
	require.NoError(t, err)
	_, err = fs.Create("/some/directory//subfolder/some/chart/templates/some-file.yaml")
	require.NoError(t, err)
	_, err = fs.Create("/some/directory//subfolder/some/chart/templates/_helpers.tpl")
	require.NoError(t, err)
	_, err = fs.Create("/some/directory//subfolder/some/chart/tests/test-1.yaml")
	require.NoError(t, err)
	_, err = fs.Create("/some/directory//subfolder/some/chart/tests/some/details/test-2.yaml")
	require.NoError(t, err)
	testValues := []string{}
	for d := range helm.DiscoverChartTests("/some/directory//subfolder/some/chart") {
		testValues = append(testValues, d)
	}
	assert.Len(t, testValues, 2)
	assert.Contains(t, testValues, "/some/directory/subfolder/some/chart/tests/test-1.yaml")
	assert.Contains(t, testValues, "/some/directory/subfolder/some/chart/tests/some/details/test-2.yaml")
}

func TestLoadMetadata(t *testing.T) {
	t.Cleanup(system.Reset)
	fs := afero.NewMemMapFs()
	system.DefaultFileSystem = fs
	testutils.EnsureFileContent(t, fs, "/chart/Chart.yaml", `{"name": "chart-name"}`)
	metadata, err := helm.LoadMetadata("/chart")
	require.NoError(t, err)
	assert.Equal(t, "chart-name", metadata.Name)
}
