package testutils

import (
	helm "github.com/adevinta/go-helm-toolkit"
)

var (
	originalDefaultHelm = defaultHelm
)

func ResetHooks() {
	defaultHelm = originalDefaultHelm
}

func SetDefaultHelm(f func() (helm.Helm, error)) {
	defaultHelm = f
}
