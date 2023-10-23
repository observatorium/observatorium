package receive

import "gopkg.in/yaml.v2"

// Taken from https://github.com/thanos-io/thanos/blob/release-0.32/pkg/receive/limiter_config.go

// ReceiveLimitsConfig is the root configuration for limits.
type ReceiveLimitsConfig struct {
	// WriteLimits hold the limits for writing data.
	WriteLimits WriteLimitsConfig `yaml:"write,omitempty"`
}

func NewReceiveLimitsConfig() ReceiveLimitsConfig {
	return ReceiveLimitsConfig{}
}

// String returns a string representation of the RootLimitsConfig as JSON.
// It implements the Stringer interface that is used by the cmdopt package.
func (r ReceiveLimitsConfig) String() string {
	// We use "gopkg.in/yaml.v2" instead of "github.com/ghodss/yaml" for correct formatting of this config.
	ret, err := yaml.Marshal(r)
	if err != nil {
		panic(err)
	}

	return string(ret)
}

func (r ReceiveLimitsConfig) WithGlobalLimits(concurrency int, monUrl, query string) ReceiveLimitsConfig {
	r.WriteLimits.GlobalLimits.MaxConcurrency = concurrency
	r.WriteLimits.GlobalLimits.MetaMonitoringURL = monUrl
	r.WriteLimits.GlobalLimits.MetaMonitoringLimitQuery = query
	return r
}

func (r ReceiveLimitsConfig) WithDefaultLimits(sizeBytes, series, samples int, headSeries int) ReceiveLimitsConfig {
	r.WriteLimits.DefaultLimits.RequestLimits.SizeBytesLimit = sizeBytes
	r.WriteLimits.DefaultLimits.RequestLimits.SeriesLimit = series
	r.WriteLimits.DefaultLimits.RequestLimits.SamplesLimit = samples
	r.WriteLimits.DefaultLimits.HeadSeriesLimit = headSeries
	return r
}

func (r ReceiveLimitsConfig) WithTenantLimits(tenant string, sizeBytes, series, samples int, headSeries uint64) ReceiveLimitsConfig {
	r.WriteLimits.TenantsLimits[tenant] = WriteLimitConfig{
		RequestLimits: RequestLimitsConfig{
			SizeBytesLimit: sizeBytes,
			SeriesLimit:    series,
			SamplesLimit:   samples,
		},
		HeadSeriesLimit: headSeries,
	}
	return r
}

type WriteLimitsConfig struct {
	// GlobalLimits are limits that are shared across all tenants.
	GlobalLimits GlobalLimitsConfig `yaml:"global,omitempty"`
	// DefaultLimits are the default limits for tenants without specified limits.
	DefaultLimits DefaultLimitsConfig `yaml:"default,omitempty"`
	// TenantsLimits are the limits per tenant.
	TenantsLimits TenantsWriteLimitsConfig `yaml:"tenants,omitempty"`
}

type GlobalLimitsConfig struct {
	// MaxConcurrency represents the maximum concurrency during write operations.
	MaxConcurrency int `yaml:"max_concurrency,omitempty"`
	// MetaMonitoring options specify the query, url and client for Query API address used in head series limiting.
	MetaMonitoringURL        string `yaml:"meta_monitoring_url,omitempty"`
	MetaMonitoringLimitQuery string `yaml:"meta_monitoring_limit_query,omitempty"`
}

type DefaultLimitsConfig struct {
	// RequestLimits holds the difficult per-request limits.
	RequestLimits RequestLimitsConfig `yaml:"request,omitempty"`
	// HeadSeriesLimit specifies the maximum number of head series allowed for any tenant.
	HeadSeriesLimit int `yaml:"head_series_limit,omitempty"`
}

// TenantsWriteLimitsConfig is a map of tenant IDs to their WriteLimitConfig.
type TenantsWriteLimitsConfig map[string]WriteLimitConfig

// A tenant might not always have limits configured, so things here must
// use pointers.
type WriteLimitConfig struct {
	// RequestLimits holds the difficult per-request limits.
	RequestLimits RequestLimitsConfig `yaml:"request,omitempty"`
	// HeadSeriesLimit specifies the maximum number of head series allowed for a tenant.
	HeadSeriesLimit uint64 `yaml:"head_series_limit,omitempty"`
}

type RequestLimitsConfig struct {
	SizeBytesLimit int `yaml:"size_bytes_limit,omitempty"`
	SeriesLimit    int `yaml:"series_limit,omitempty"`
	SamplesLimit   int `yaml:"samples_limit,omitempty"`
}
