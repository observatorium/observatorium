package objstore

// Taken from github.com/thanos-io/objstore/client/factory.go eb06103887ab787f47d08e8a2f100264087319d5

type BucketConfig struct {
	Type   ObjProvider `yaml:"type"`
	Config interface{} `yaml:"config"`
	Prefix string      `yaml:"prefix,omitempty"`
}

type ObjProvider string

const (
	FILESYSTEM ObjProvider = "FILESYSTEM"
	GCS        ObjProvider = "GCS"
	S3         ObjProvider = "S3"
	AZURE      ObjProvider = "AZURE"
	SWIFT      ObjProvider = "SWIFT"
	COS        ObjProvider = "COS"
	ALIYUNOSS  ObjProvider = "ALIYUNOSS"
	BOS        ObjProvider = "BOS"
	OCI        ObjProvider = "OCI"
	OBS        ObjProvider = "OBS"
)
