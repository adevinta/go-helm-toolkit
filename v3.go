package helm

import (
	"io"
	"io/ioutil"
)

type Helm3 struct {
	Path        string
	APIVersions []string
	GlobalFlags []string
}

func (h *Helm3) Template(namespace, release, chart string, flags ...Flag) (io.Reader, error) {
	args := append(h.GlobalFlags, "template", release, chart, "--namespace", namespace)
	for _, flag := range flags {
		args = append(args, flag()...)
	}
	for _, v := range h.APIVersions {
		args = append(args, "--api-versions", v)
	}
	return getCommandOutput(h.Path, args...)
}

func (h *Helm3) Install(namespace, release, chart string, flags ...Flag) error {
	args := append(h.GlobalFlags, "install", release, chart, "--namespace", namespace)
	for _, flag := range flags {
		args = append(args, flag()...)
	}
	return runCommand(h.Path, args...)
}

func (h *Helm3) Update(namespace, release, chart string, flags ...Flag) error {
	args := append(h.GlobalFlags, "upgrade", release, chart, "--namespace", namespace)
	for _, flag := range flags {
		args = append(args, flag()...)
	}
	return runCommand(h.Path, args...)
}

func (h *Helm3) UpdateDeps(chart string) error {
	args := append(h.GlobalFlags, "dependency", "update", chart)
	return runCommand(h.Path, args...)
}

func (h *Helm3) Test(namespace, release string) error {
	args := append(h.GlobalFlags, "test", release, "--namespace", namespace)
	return runCommand(h.Path, args...)
}

func (h *Helm3) Delete(namespace, release string) error {
	args := append(h.GlobalFlags, "delete", release, "--namespace", namespace)
	return runCommand(h.Path, args...)
}

func (h *Helm3) Package(chart string, flags ...Flag) error {
	args := append(h.GlobalFlags, "package", chart)
	for _, flag := range flags {
		args = append(args, flag()...)
	}
	return runCommand(h.Path, args...)
}

func (h *Helm3) Init() error {
	return nil
}

func (h *Helm3) RepoAdd(name, url string, flags ...Flag) error {
	args := append(h.GlobalFlags, "repo", "add", name, url)
	for _, flag := range flags {
		args = append(args, flag()...)
	}
	return runCommand(h.Path, args...)
}
func (h *Helm3) RepoUpdate() error {
	args := append(h.GlobalFlags, "repo", "update")
	return runCommand(h.Path, args...)
}

func (h *Helm3) Version() string {
	args := append(h.GlobalFlags, "version", "--template", "{{.Version}}")
	reader, err := getCommandOutput(h.Path, args...)
	if err != nil {
		return ""
	}
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return ""
	}
	return string(b)
}
