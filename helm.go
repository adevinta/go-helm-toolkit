package helm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	goruntime "runtime"

	system "github.com/adevinta/go-system-toolkit"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

const DefaultHelmVersion = "v3.18.4"

type Flag func() []string

func Debug() Flag {
	return func() []string {
		return []string{"--debug"}
	}
}

func UpgradeInstall() Flag {
	return func() []string {
		return []string{"--install"}
	}
}

func Values(path string) Flag {
	return func() []string {
		return []string{"--values", path}
	}
}

func RepoUsername(username string) Flag {
	return func() []string {
		return []string{"--username", username}
	}
}

func RepoPassword(password string) Flag {
	return func() []string {
		return []string{"--password", password}
	}
}

func Version(v string) Flag {
	return func() []string {
		return []string{"--version", v}
	}
}

type Helm interface {
	Template(namespace, release, chart string, flags ...Flag) (io.Reader, error)
	Install(namespace, release, chart string, flags ...Flag) error
	Update(namespace, release, chart string, flags ...Flag) error
	Package(chart string, flags ...Flag) error
	UpdateDeps(chart string) error
	Test(namespace, release string) error
	Delete(namespace, release string) error
	Init() error
	RepoAdd(name, url string, flags ...Flag) error
	RepoUpdate() error
	Version() string
}

func Download(version, dest string) error {
	resp, err := http.Get(fmt.Sprintf("https://get.helm.sh/helm-%s-%s-%s.tar.gz", version, goruntime.GOOS, goruntime.GOARCH))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	tarBody, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(tarBody)
	found := false
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			if found {
				return nil
			}
			return errors.New("Unable to find helm binary")
		}
		if header.Typeflag == tar.TypeReg && strings.HasSuffix(header.Name, "/helm") {
			found = true
			if err := os.MkdirAll(filepath.Dir(dest), 0777); err != nil {
				return err
			}
			fd, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			defer fd.Close()
			_, err = io.Copy(fd, tarReader)
			if err != nil {
				return err
			}
		}
	}

}

func envOrDefault(key, dflt string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return dflt
	}
	return val
}

func Default() (Helm, error) {
	path := ".helm/bin/helm-" + DefaultHelmVersion
	err := Download(DefaultHelmVersion, path)
	if err != nil {
		return nil, err
	}
	return &Helm3{
		Path:        path,
		APIVersions: strings.Split(envOrDefault("HELM3_API_VERSIONS", "1.18.0"), ","),
	}, nil
}

func Supported() <-chan Helm {
	o := make(chan Helm)
	go func() {
		for _, v := range strings.Split(envOrDefault("HELM3_VERSIONS", DefaultHelmVersion), ",") {
			err := Download(v, ".helm/bin/helm-"+v)
			if err == nil {
				o <- &Helm3{
					Path:        ".helm/bin/helm-" + v,
					APIVersions: strings.Split(envOrDefault("HELM3_API_VERSIONS", "1.18.0"), ","),
				}
			}
		}
		close(o)
	}()
	return o
}

func getCommandOutput(cmd string, args ...string) (io.Reader, error) {
	b := &bytes.Buffer{}
	c := exec.Command(cmd, args...)
	c.Stdout = b
	c.Stderr = os.Stderr
	err := c.Run()
	return b, err
}

func runCommand(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func DiscoverChartDirs(path string) <-chan string {
	o := make(chan string)
	go func() {
		afero.Walk(system.DefaultFileSystem, path, func(path string, info os.FileInfo, err error) error {
			if filepath.Base(path) == "Chart.yaml" {
				o <- filepath.Dir(path)
			}
			return nil
		})
		close(o)
	}()
	return o
}

func DiscoverChartTests(path string) <-chan string {
	o := make(chan string)
	go func() {
		afero.Walk(system.DefaultFileSystem, filepath.Join(path, "tests"), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".yaml") {
				o <- path
			}
			return nil
		})
		close(o)
	}()
	return o
}

func LoadMetadata(path string) (*Metadata, error) {
	fd, err := system.DefaultFileSystem.Open(filepath.Join(path, "Chart.yaml"))
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	meta := Metadata{}
	return &meta, yaml.NewDecoder(fd).Decode(&meta)
}
