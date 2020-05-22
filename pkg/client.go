package insights

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type authType string

const (
	authTypeBasic authType = "basic"
	authTypeCert           = "cert"
)

// Client is a specialized HTTP client, preconfigured to authenticate using
// either certificate or basic authentication.
type Client struct {
	*http.Client
	authType authType
	username string
	password string
	baseURL  string
}

// NewClientBasicAuth creates a client configured for basic authentication with
// the given username and password.
func NewClientBasicAuth(baseURL, username, password string) (*Client, error) {
	return &Client{
		Client:   &http.Client{},
		authType: authTypeBasic,
		baseURL:  baseURL,
		username: username,
		password: password,
	}, nil
}

// NewClientCertAuth creates a client configured for certificate authentication
// with the given CA root, and certificate key-pair.
func NewClientCertAuth(baseURL, caRoot, certFile, keyFile string) (*Client, error) {
	client := &Client{
		Client:   &http.Client{},
		authType: authTypeCert,
		baseURL:  baseURL,
	}

	tlsConfig := tls.Config{
		MaxVersion: tls.VersionTLS12, // cloud.redhat.com appears to exhibit this openssl bug https://github.com/openssl/openssl/issues/9767
	}

	caCert, err := ioutil.ReadFile(caRoot)
	if err != nil {
		return nil, err
	}
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig.RootCAs = caCertPool

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}
	tlsConfig.BuildNameToCertificate()

	client.Transport = &http.Transport{
		TLSClientConfig: &tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return client, nil
}

// Do sends an HTTP request and returns an HTTP response, following policy
// as configured on the client.
//
// See http.Client documentation for more details.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.authType == authTypeBasic {
		req.SetBasicAuth(c.username, c.password)
	}
	req.Header.Add("User-Agent", "insightsd/1")
	return c.Client.Do(req)
}

func (c *Client) doReq(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) Get(url string) (*http.Response, error) {
	return c.doReq(http.MethodGet, url, nil)
}

func (c *Client) Put(url string, body io.Reader) (*http.Response, error) {
	return c.doReq(http.MethodPut, url, body)
}

func (c *Client) Post(url string, body io.Reader) (*http.Response, error) {
	return c.doReq(http.MethodPost, url, body)
}
