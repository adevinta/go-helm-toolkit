package helm_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	pkggodevclient "github.com/guseggert/pkggodev-client"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/modfile"
	"golang.org/x/time/rate"
)

var (
	acceptedLicenses = map[string]struct{}{
		"MIT":          {},
		"Apache-2.0":   {},
		"BSD-3-Clause": {},
		"BSD-2-Clause": {},
		"ISC":          {},
		"CC-BY-SA-4.0": {},
		"MPL-2.0":      {},
	}

	knownUndectedLicenses = map[string]string{
		// bufpipe was later added the MIT license: https://github.com/acomagu/bufpipe/blob/cd7a5f79d3c413d14c0c60fd31dae7b397fc955a/LICENSE
		"github.com/acomagu/bufpipe@v1.0.3": "MIT",
	}
)

func TestLicenses(t *testing.T) {

	// Create HTTP client with throttling and backoff retry
	limiter := rate.NewLimiter(rate.Every(1*time.Second), 1) // 1 request per second with burst of 1

	httpClient := &http.Client{
		Timeout: 180 * time.Second,
		Transport: &throttledTransport{
			limiter: limiter,
			base:    http.DefaultTransport,
		},
	}

	b, err := ioutil.ReadFile("go.mod")
	require.NoError(t, err)
	file, err := modfile.Parse("go.mod", b, nil)
	require.NoError(t, err)

	client := pkggodevclient.New(pkggodevclient.WithHTTPClient(httpClient))
	for _, req := range file.Require {
		pkg, err := client.DescribePackage(pkggodevclient.DescribePackageRequest{
			Package: req.Mod.Path,
		})
		require.NoError(t, err)
		licences := strings.Split(pkg.License, ",")
		for _, license := range licences {
			license = strings.TrimSpace(license)
			if license == "None detected" {
				if known, ok := knownUndectedLicenses[req.Mod.String()]; ok {
					license = known
				}
			}
			if _, ok := acceptedLicenses[license]; !ok {
				t.Errorf("dependency %s is using unexpected license %s. Check that this license complies with MIT in which maiao is released and update the checks accordingly or change dependency", req.Mod, license)
			}
		}
	}
}

// throttledTransport wraps an http.RoundTripper with rate limiting
type throttledTransport struct {
	limiter *rate.Limiter
	base    http.RoundTripper
}

func (t *throttledTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Wait for rate limiter
	err := t.limiter.Wait(context.Background())
	if err != nil {
		return nil, err
	}

	// Try the request with exponential backoff for rate limiting
	var resp *http.Response
	var lastErr error
	maxRetries := 5
	baseDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		resp, lastErr = t.base.RoundTrip(req)
		if lastErr != nil {
			return nil, lastErr
		}

		// If we get a 429 (Too Many Requests) or 503 (Service Unavailable), retry with exponential backoff
		if resp.StatusCode == 429 || resp.StatusCode == 503 {
			if i < maxRetries-1 { // Don't sleep on the last attempt
				// Exponential backoff: 2s, 4s, 8s, 16s, 32s
				delay := baseDelay * time.Duration(1<<i)
				// Add some jitter to avoid thundering herd
				jitter := time.Duration(float64(delay) * 0.1 * (0.5 + 0.5*float64(i)))
				totalDelay := delay + jitter

				fmt.Printf("Rate limited (HTTP %d), retrying in %v (attempt %d/%d)\n", resp.StatusCode, totalDelay, i+1, maxRetries)
				time.Sleep(totalDelay)
				resp.Body.Close()
				continue
			} else {
				fmt.Printf("Rate limited (HTTP %d), giving up after %d attempts\n", resp.StatusCode, maxRetries)
			}
		}

		return resp, nil
	}

	return resp, lastErr
}
