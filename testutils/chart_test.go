package testutils_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	helm "github.com/adevinta/go-helm-toolkit"
	helmtestutils "github.com/adevinta/go-helm-toolkit/testutils"
	k8s "github.com/adevinta/go-k8s-toolkit"
	k8stestutils "github.com/adevinta/go-k8s-toolkit/testutils"
	testutils "github.com/adevinta/go-testutils-toolkit"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type helmTemplateFunc struct {
	helm.Helm
	templateFunc  func(namespace, release, chart string, flags ...helm.Flag) (io.Reader, error)
	templateCalls int
}

func (h *helmTemplateFunc) Template(namespace, release, chart string, flags ...helm.Flag) (io.Reader, error) {
	h.templateCalls++
	if h.templateFunc != nil {
		return h.templateFunc(namespace, release, chart, flags...)
	}
	if h.Helm != nil {
		return h.Helm.Template(namespace, release, chart, flags...)
	}
	return nil, errors.New("Template not available, templateFunc and Helm are nil")
}

func TestInstallFilteredHelmChart(t *testing.T) {
	t.Cleanup(helmtestutils.ResetHooks)
	ft := &testutils.FakeTest{}
	helmtestutils.SetDefaultHelm(func() (helm.Helm, error) {
		return &helmTemplateFunc{
			templateFunc: func(namespace, release, chart string, flags ...helm.Flag) (io.Reader, error) {
				assert.Equal(t, "my-namespace", namespace)
				assert.Equal(t, "test-release", release)
				assert.Equal(t, "chart-name", chart)
				args := []string{}
				for _, f := range flags {
					args = append(args, f()...)
				}
				assert.Equal(t, []string{"--include-crds"}, args)
				b := &bytes.Buffer{}
				k8s.SerialiseObjects(
					scheme.Scheme,
					b,
					&v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-namespace",
						},
					},
					&v1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-cm",
							Namespace: "my-namespace",
						},
					},
					&v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "my-secret",
							Namespace: "my-namespace",
						},
					},
				)
				return b, nil
			},
		}, nil
	})

	c := fake.NewClientBuilder().Build()
	helmtestutils.InstallFilteredHelmChart(ft, context.Background(), c, "my-namespace", "test-release", "chart-name", k8stestutils.ExcludeObject(k8stestutils.WithKind("ConfigMap")))
	assert.False(t, ft.Failed)
	assert.Len(t, ft.ErrorMessages, 0)

	k8stestutils.AssertHasObject(t, c, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-secret",
			Namespace: "my-namespace",
		},
	})
	k8stestutils.AssertHasObject(t, c, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-namespace",
		},
	})
	k8stestutils.AssertHasNoObject(t, c, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cm",
			Namespace: "my-namespace",
		},
	})
}
