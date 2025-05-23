/*
   Copyright The Soci Snapshotter Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package resolver

// =====
// This is CRI-plugin-compatible registry hosts configuration.
// Some functions are ported from https://github.com/containerd/containerd/blob/v1.5.2/pkg/cri as noted on each one.
// TODO: import them from CRI package once we drop support to continerd v1.4.x
// =====

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/reference"
	"github.com/containerd/containerd/remotes/docker"
	dconfig "github.com/containerd/containerd/remotes/docker/config"
	runtime_alpha "github.com/containerd/containerd/third_party/k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"github.com/containerd/errdefs"
	rhttp "github.com/hashicorp/go-retryablehttp"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// Registry is registry settings configured
type Registry struct {
	// ConfigPath is a path to the root directory containing registry-specific
	// configurations.
	// If ConfigPath is set, the rest of the registry specific options are ignored.
	ConfigPath string `toml:"config_path" json:"configPath"`
	// Mirrors are namespace to mirror mapping for all namespaces.
	// This option will not be used when ConfigPath is provided.
	// DEPRECATED: Use ConfigPath instead. Remove in containerd 1.7.
	Mirrors map[string]Mirror `toml:"mirrors" json:"mirrors"`
	// Configs are configs for each registry.
	// The key is the domain name or IP of the registry.
	// This option will be fully deprecated for ConfigPath in the future.
	Configs map[string]RegistryConfig `toml:"configs" json:"configs"`
}

// Mirror contains the config related to the registry mirror
type Mirror struct {
	// Endpoints are endpoints for a namespace. CRI plugin will try the endpoints
	// one by one until a working one is found. The endpoint must be a valid url
	// with host specified.
	// The scheme, host and path from the endpoint URL will be used.
	Endpoints []string `toml:"endpoint" json:"endpoint"`
}

// RegistryConfig contains configuration used to communicate with the registry.
type RegistryConfig struct {
	// Auth contains information to authenticate to the registry.
	Auth *AuthConfig `toml:"auth" json:"auth"`
	// TLS is a pair of CA/Cert/Key which then are used when creating the transport
	// that communicates with the registry.
	// This field will not be used when ConfigPath is provided.
	// DEPRECATED: Use ConfigPath instead. Remove in containerd 1.7.
	TLS *TLSConfig `toml:"tls" json:"tls"`
}

// AuthConfig contains the config related to authentication to a specific registry
type AuthConfig struct {
	// Username is the username to login the registry.
	Username string `toml:"username" json:"username"`
	// Password is the password to login the registry.
	Password string `toml:"password" json:"password"`
	// Auth is a base64 encoded string from the concatenation of the username,
	// a colon, and the password.
	Auth string `toml:"auth" json:"auth"`
	// IdentityToken is used to authenticate the user and get
	// an access token for the registry.
	IdentityToken string `toml:"identitytoken" json:"identitytoken"`
}

// TLSConfig contains the CA/Cert/Key used for a registry
type TLSConfig struct {
	InsecureSkipVerify bool   `toml:"insecure_skip_verify" json:"insecure_skip_verify"`
	CAFile             string `toml:"ca_file" json:"caFile"`
	CertFile           string `toml:"cert_file" json:"certFile"`
	KeyFile            string `toml:"key_file" json:"keyFile"`
}

// RegistryHostsFromCRIConfig creates RegistryHosts (a set of registry configuration) from CRI-plugin-compatible config.
// NOTE: ported from https://github.com/containerd/containerd/blob/v1.5.2/pkg/cri/server/image_pull.go#L332-L405
// TODO: import this from CRI package once we drop support to continerd v1.4.x
func RegistryHostsFromCRIConfig(ctx context.Context, config Registry, credsFuncs ...Credential) RegistryHosts {
	paths := filepath.SplitList(config.ConfigPath)
	if len(paths) > 0 {
		return func(imgRefSpec reference.Spec) ([]docker.RegistryHost, error) {
			host := imgRefSpec.Hostname()
			hostOptions := dconfig.HostOptions{}
			hostOptions.Credentials = multiCredsFuncs(imgRefSpec, append(credsFuncs, func(imgRefSpec reference.Spec, host string) (string, string, error) {
				config := config.Configs[host]
				if config.Auth != nil {
					return ParseAuth(toRuntimeAuthConfig(*config.Auth), host)
				}
				return "", "", nil
			})...)
			hostOptions.HostDir = hostDirFromRoots(paths)
			return dconfig.ConfigureHosts(ctx, hostOptions)(host)
		}
	}
	return func(imgRefSpec reference.Spec) ([]docker.RegistryHost, error) {
		host := imgRefSpec.Hostname()
		var registries []docker.RegistryHost
		endpoints, err := registryEndpoints(config, host)
		if err != nil {
			return nil, fmt.Errorf("get registry endpoints: %w", err)
		}
		for _, e := range endpoints {
			u, err := url.Parse(e)
			if err != nil {
				return nil, fmt.Errorf("parse registry endpoint %q from mirrors: %w", e, err)
			}

			var (
				rclient = rhttp.NewClient()
				config  = config.Configs[u.Host]
			)

			rclient.Logger = nil // disable logging every request

			if config.TLS != nil {
				if tr, ok := rclient.HTTPClient.Transport.(*http.Transport); ok {
					tr.TLSClientConfig, err = getTLSConfig(*config.TLS)
					if err != nil {
						return nil, fmt.Errorf("get TLSConfig for registry %q: %w", e, err)
					}
				} else {
					return nil, errors.New("TLS config cannot be applied; Client.Transport is not *http.Transport")
				}
			}

			client := rclient.StandardClient()
			authorizer := docker.NewDockerAuthorizer(
				docker.WithAuthClient(client),
				docker.WithAuthCreds(multiCredsFuncs(imgRefSpec, credsFuncs...)))

			if u.Path == "" {
				u.Path = "/v2"
			}

			registries = append(registries, docker.RegistryHost{
				Client:       client,
				Authorizer:   authorizer,
				Host:         u.Host,
				Scheme:       u.Scheme,
				Path:         u.Path,
				Capabilities: docker.HostCapabilityResolve | docker.HostCapabilityPull,
			})
		}
		return registries, nil
	}
}

// Ported from https://github.com/containerd/containerd/blob/v1.5.2/pkg/cri/server/image_pull.go#L316-L330
// TODO: import this from CRI package once we drop support to continerd v1.4.x
func hostDirFromRoots(roots []string) func(string) (string, error) {
	rootfn := make([]func(string) (string, error), len(roots))
	for i := range roots {
		rootfn[i] = dconfig.HostDirFromRoot(roots[i])
	}
	return func(host string) (dir string, err error) {
		for _, fn := range rootfn {
			dir, err = fn(host)
			if (err != nil && !errdefs.IsNotFound(err)) || (dir != "") {
				break
			}
		}
		return
	}
}

// toRuntimeAuthConfig converts cri plugin auth config to runtime auth config.
// Ported from https://github.com/containerd/containerd/blob/v1.5.2/pkg/cri/server/helpers.go#L295-L303
// TODO: import this from CRI package once we drop support to continerd v1.4.x
func toRuntimeAuthConfig(a AuthConfig) *runtime.AuthConfig {
	return &runtime.AuthConfig{
		Username:      a.Username,
		Password:      a.Password,
		Auth:          a.Auth,
		IdentityToken: a.IdentityToken,
	}
}

// getTLSConfig returns a TLSConfig configured with a CA/Cert/Key specified by registryTLSConfig
// Ported from https://github.com/containerd/containerd/blob/v1.5.2/pkg/cri/server/image_pull.go#L316-L330
// TODO: import this from CRI package once we drop support to continerd v1.4.x
func getTLSConfig(registryTLSConfig TLSConfig) (*tls.Config, error) {
	var (
		tlsConfig = &tls.Config{}
		cert      tls.Certificate
		err       error
	)
	if registryTLSConfig.CertFile != "" && registryTLSConfig.KeyFile == "" {
		return nil, fmt.Errorf("cert file %q was specified, but no corresponding key file was specified", registryTLSConfig.CertFile)
	}
	if registryTLSConfig.CertFile == "" && registryTLSConfig.KeyFile != "" {
		return nil, fmt.Errorf("key file %q was specified, but no corresponding cert file was specified", registryTLSConfig.KeyFile)
	}
	if registryTLSConfig.CertFile != "" && registryTLSConfig.KeyFile != "" {
		cert, err = tls.LoadX509KeyPair(registryTLSConfig.CertFile, registryTLSConfig.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load cert file: %w", err)
		}
		if len(cert.Certificate) != 0 {
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		tlsConfig.BuildNameToCertificate() // nolint:staticcheck
	}

	if registryTLSConfig.CAFile != "" {
		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to get system cert pool: %w", err)
		}
		caCert, err := os.ReadFile(registryTLSConfig.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA file: %w", err)
		}
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	tlsConfig.InsecureSkipVerify = registryTLSConfig.InsecureSkipVerify
	return tlsConfig, nil
}

// addDefaultScheme returns the endpoint with default scheme
// Ported from https://github.com/containerd/containerd/blob/v1.5.2/pkg/cri/server/image_pull.go#L316-L330
// TODO: import this from CRI package once we drop support to continerd v1.4.x
func addDefaultScheme(endpoint string) (string, error) {
	if strings.Contains(endpoint, "://") {
		return endpoint, nil
	}
	ue := "dummy://" + endpoint
	u, err := url.Parse(ue)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s://%s", DefaultScheme(u.Host), endpoint), nil
}

// registryEndpoints returns endpoints for a given host.
// It adds default registry endpoint if it does not exist in the passed-in endpoint list.
// It also supports wildcard host matching with `*`.
// Ported from https://github.com/containerd/containerd/blob/v1.5.2/pkg/cri/server/image_pull.go#L431-L464
// TODO: import this from CRI package once we drop support to continerd v1.4.x
func registryEndpoints(config Registry, host string) ([]string, error) {
	var endpoints []string
	_, ok := config.Mirrors[host]
	if ok {
		endpoints = config.Mirrors[host].Endpoints
	} else {
		endpoints = config.Mirrors["*"].Endpoints
	}
	defaultHost, err := docker.DefaultHost(host)
	if err != nil {
		return nil, fmt.Errorf("get default host: %w", err)
	}
	for i := range endpoints {
		en, err := addDefaultScheme(endpoints[i])
		if err != nil {
			return nil, fmt.Errorf("parse endpoint url: %w", err)
		}
		endpoints[i] = en
	}
	for _, e := range endpoints {
		u, err := url.Parse(e)
		if err != nil {
			return nil, fmt.Errorf("parse endpoint url: %w", err)
		}
		if u.Host == host {
			// Do not add default if the endpoint already exists.
			return endpoints, nil
		}
	}
	return append(endpoints, DefaultScheme(defaultHost)+"://"+defaultHost), nil
}

// ParseAlphaAuth parses AuthConfig and returns username and password/secret required by containerd.
// Ported from https://github.com/containerd/containerd/blob/v1.5.2/pkg/cri/server/image_pull.go#L176-L214
// TODO: import this from CRI package once we drop support to continerd v1.4.x
func ParseAlphaAuth(auth *runtime_alpha.AuthConfig, host string) (string, string, error) {
	if auth == nil {
		return "", "", nil
	}
	if auth.ServerAddress != "" {
		// Do not return the auth info when server address doesn't match.
		u, err := url.Parse(auth.ServerAddress)
		if err != nil {
			return "", "", fmt.Errorf("parse server address: %w", err)
		}
		if host != u.Host {
			return "", "", nil
		}
	}
	if auth.Username != "" {
		return auth.Username, auth.Password, nil
	}
	if auth.IdentityToken != "" {
		return "", auth.IdentityToken, nil
	}
	if auth.Auth != "" {
		decLen := base64.StdEncoding.DecodedLen(len(auth.Auth))
		decoded := make([]byte, decLen)
		_, err := base64.StdEncoding.Decode(decoded, []byte(auth.Auth))
		if err != nil {
			return "", "", err
		}
		fields := strings.SplitN(string(decoded), ":", 2)
		if len(fields) != 2 {
			return "", "", fmt.Errorf("invalid decoded auth: %q", decoded)
		}
		user, passwd := fields[0], fields[1]
		return user, strings.Trim(passwd, "\x00"), nil
	}
	// TODO(random-liu): Support RegistryToken.
	// An empty auth config is valid for anonymous registry
	return "", "", nil
}

// ParseAuth parses AuthConfig and returns username and password/secret required by containerd.
// Ported from https://github.com/containerd/containerd/blob/v1.5.2/pkg/cri/server/image_pull.go#L176-L214
// TODO: import this from CRI package once we drop support to continerd v1.4.x
func ParseAuth(auth *runtime.AuthConfig, host string) (string, string, error) {
	if auth == nil {
		return "", "", nil
	}
	if auth.ServerAddress != "" {
		// Do not return the auth info when server address doesn't match.
		u, err := url.Parse(auth.ServerAddress)
		if err != nil {
			return "", "", fmt.Errorf("parse server address: %w", err)
		}
		if host != u.Host {
			return "", "", nil
		}
	}
	if auth.Username != "" {
		return auth.Username, auth.Password, nil
	}
	if auth.IdentityToken != "" {
		return "", auth.IdentityToken, nil
	}
	if auth.Auth != "" {
		decLen := base64.StdEncoding.DecodedLen(len(auth.Auth))
		decoded := make([]byte, decLen)
		_, err := base64.StdEncoding.Decode(decoded, []byte(auth.Auth))
		if err != nil {
			return "", "", err
		}
		fields := strings.SplitN(string(decoded), ":", 2)
		if len(fields) != 2 {
			return "", "", fmt.Errorf("invalid decoded auth: %q", decoded)
		}
		user, passwd := fields[0], fields[1]
		return user, strings.Trim(passwd, "\x00"), nil
	}
	// TODO(random-liu): Support RegistryToken.
	// An empty auth config is valid for anonymous registry
	return "", "", nil
}
