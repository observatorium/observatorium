package api

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v2"
)

// Taken from https://github.com/observatorium/api/blob/ce3e8a59e994ad2798c218a432afd37213ed8459/main.go#L199

type Tenants struct {
	Tenants []Tenant `yaml:"tenants"`
}

func (t Tenants) String() string {
	res, err := yaml.Marshal(t)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal tenants file: %v", err))
	}
	return string(res)
}

type Tenant struct {
	Name          string               `yaml:"name"`
	ID            string               `yaml:"id"`
	OIDC          *TenantOIDC          `yaml:"oidc,omitempty"`
	OpenShift     *TenantOpenShift     `yaml:"openshift,omitempty"`
	Authenticator *TenantAuthenticator `yaml:"authenticator,omitempty"`
	MTLS          *TenantMTLS          `yaml:"mTLS,omitempty"`
	OPA           *TenantOPA           `yaml:"opa,omitempty"`
	RateLimits    []TenantRateLimits   `yaml:"rateLimits,omitempty"`
}

type TenantOIDC struct {
	ClientID      string `yaml:"clientID"`
	ClientSecret  string `yaml:"clientSecret"`
	GroupClaim    string `yaml:"groupClaim,omitempty"`
	IssuerRawCA   []byte `yaml:"issuerCA,omitempty"`
	IssuerCAPath  string `yaml:"issuerCAPath,omitempty"`
	IssuerURL     string `yaml:"issuerURL"`
	RedirectURL   string `yaml:"redirectURL"`
	UsernameClaim string `yaml:"usernameClaim,omitempty"`
}

type TenantOpenShift struct {
	KubeConfigPath string `yaml:"kubeconfig"`
	ServiceAccount string `yaml:"serviceAccount"`
	RedirectURL    string `yaml:"redirectURL"`
	CookieSecret   string `yaml:"cookieSecret"`
}

type TenantAuthenticator struct {
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

type TenantMTLS struct {
	RawCA  []byte `yaml:"ca"`
	CAPath string `yaml:"caPath"`
}

type TenantOPA struct {
	Query           string   `yaml:"query,omitempty"`
	Paths           []string `yaml:"paths,omitempty"`
	URL             string   `yaml:"url"`
	WithAccessToken bool     `yaml:"withAccessToken,omitempty"`
}

type TenantRateLimits struct {
	Endpoint string        `yaml:"endpoint"`
	Limit    int           `yaml:"limit"`
	Window   time.Duration `yaml:"window"`
	// The remaining fields in this struct are optional and only apply to the remote rate limiter.
	// FailOpen determines the behavior of the rate limiter when a remote rate limiter is unavailable.
	// If true, requests will be accepted when the remote rate limiter decision is unavailable or returns an error.
	FailOpen bool `yaml:"failOpen,omitempty"`
	// RetryAfterMin and RetryAfterMax are used to determine the Retry-After header value when the
	// remote rate limiter determines that the request should be rejected.
	// This can be used to prevent a thundering herd of requests from overwhelming the upstream and is
	// respected by the Prometheus remote write client.
	// As requests get rejected the header is set and the value doubled each time until RetryAfterMaxSeconds.
	// Zero or unset values will result in no Retry-After header being set.
	// RetryAfterMin is the minimum value for the Retry-After header.
	RetryAfterMin time.Duration `yaml:"retryAfterMin,omitempty,omitempty"`
	// RetryAfterMax is the maximum value for the Retry-After header.
	// If RetryAfterMax is zero and RetryAfterMin is non-zero, the Retry-After header will grow indefinitely.
	RetryAfterMax time.Duration `yaml:"retryAfterMax,omitempty"`
}
