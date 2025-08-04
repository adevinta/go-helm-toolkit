package testutils

import (
	"context"

	helm "github.com/adevinta/go-helm-toolkit"
	k8s "github.com/adevinta/go-k8s-toolkit"
	k8stestutils "github.com/adevinta/go-k8s-toolkit/testutils"
	testutils "github.com/adevinta/go-testutils-toolkit"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	defaultHelm = helm.Default
)

func InstallFilteredHelmChart(t require.TestingT, ctx context.Context, c client.Client, namespace, release, chart string, filters ...k8stestutils.Filter) {
	if h, ok := t.(testutils.TestHelper); ok {
		h.Helper()
	}
	h, err := defaultHelm()
	require.NoError(t, err)
	reader, err := h.Template(namespace, release, chart, func() []string { return []string{"--include-crds"} })
	require.NoError(t, err)
	objects, err := k8s.ParseUnstructured(reader)
	require.NoError(t, err)

	objects = k8stestutils.FilterUnstructuredObjects(
		objects,
		filters...,
	)
	k8stestutils.CreateOrUpdateAll(t, ctx, c, k8s.ToClientObject(objects)...)
}
