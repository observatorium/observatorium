package s3

import "time"

// Taken from github.com/thanos-io/objstore/providers/s3/s3.go eb06103887ab787f47d08e8a2f100264087319d5

type Config struct {
	Bucket             string            `yaml:"bucket"`
	Endpoint           string            `yaml:"endpoint"`
	Region             string            `yaml:"region,omitempty"`
	AWSSDKAuth         bool              `yaml:"aws_sdk_auth,omitempty"`
	AccessKey          string            `yaml:"access_key,omitempty"`
	Insecure           bool              `yaml:"insecure,omitempty"`
	SignatureV2        bool              `yaml:"signature_version2,omitempty"`
	SecretKey          string            `yaml:"secret_key,omitempty"`
	SessionToken       string            `yaml:"session_token,omitempty"`
	PutUserMetadata    map[string]string `yaml:"put_user_metadata,omitempty"`
	HTTPConfig         HTTPConfig        `yaml:"http_config,omitempty"`
	TraceConfig        TraceConfig       `yaml:"trace,omitempty"`
	ListObjectsVersion string            `yaml:"list_objects_versio,omitempty"`
	BucketLookupType   BucketLookupType  `yaml:"bucket_lookup_type,omitempty"`
	// PartSize used for multipart upload. Only used if uploaded object size is known and larger than configured PartSize.
	// NOTE we need to make sure this number does not produce more parts than 10 000.
	PartSize    uint64    `yaml:"part_size,omitempty"`
	SSEConfig   SSEConfig `yaml:"sse_config,omitempty"`
	STSEndpoint string    `yaml:"sts_endpoint,omitempty"`
}

// SSEConfig deals with the configuration of SSE for Minio. The following options are valid:
// KMSEncryptionContext == https://docs.aws.amazon.com/kms/latest/developerguide/services-s3.html#s3-encryption-context
type SSEConfig struct {
	Type                 string            `yaml:"type,omitempty"`
	KMSKeyID             string            `yaml:"kms_key_id,omitempty"`
	KMSEncryptionContext map[string]string `yaml:"kms_encryption_context,omitempty"`
	EncryptionKey        string            `yaml:"encryption_key,omitempty"`
}

type TraceConfig struct {
	Enable bool `yaml:"enable"`
}

// HTTPConfig stores the http.Transport configuration for the cos and s3 minio client.
type HTTPConfig struct {
	IdleConnTimeout       time.Duration `yaml:"idle_conn_timeout,omitempty"`
	ResponseHeaderTimeout time.Duration `yaml:"response_header_timeout,omitempty"`
	InsecureSkipVerify    bool          `yaml:"insecure_skip_verify,omitempty"`

	TLSHandshakeTimeout   time.Duration `yaml:"tls_handshake_timeout,omitempty"`
	ExpectContinueTimeout time.Duration `yaml:"expect_continue_timeout,omitempty"`
	MaxIdleConns          int           `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost   int           `yaml:"max_idle_conns_per_host,omitempty"`
	MaxConnsPerHost       int           `yaml:"max_conns_per_host,omitempty"`

	TLSConfig          TLSConfig `yaml:"tls_config,omitempty"`
	DisableCompression bool      `yaml:"disable_compression,omitempty"`
}

// TLSConfig configures the options for TLS connections.
type TLSConfig struct {
	// The CA cert to use for the targets.
	CAFile string `yaml:"ca_file,omitempty"`
	// The client cert file for the targets.
	CertFile string `yaml:"cert_file,omitempty"`
	// The client key file for the targets.
	KeyFile string `yaml:"key_file,omitempty"`
	// Used to verify the hostname for the targets.
	ServerName string `yaml:"server_name,omitempty"`
	// Disable target certificate validation.
	InsecureSkipVerify bool `yaml:"insecure_skip_verify,omitempty"`
}

type BucketLookupType string

const (
	BucketLookupTypeAuto          BucketLookupType = "auto"
	BucketLookupTypeVirtualHosted BucketLookupType = "virtual-hosted"
	BucketLookupTypePath          BucketLookupType = "path"
)
