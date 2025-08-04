# testutils

A helper to write more concise and self-explanatory tests.

## Examples

```
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
```