package option

// ConfigFile is a struct that represents a config file.
// It includes data needed to mount the file as a configMap or secret in a container.
type ConfigFile[T any] struct {
	Name      string
	Value     T
	mountPath string
	fileName  string
}

// NewConfigFile creates a new ConfigFile.
func NewConfigFile[T any](mountPath, fileName, name string, value T) *ConfigFile[T] {
	return &ConfigFile[T]{
		mountPath: mountPath,
		fileName:  fileName,
		Name:      name,
		Value:     value,
	}
}

// String returns the path to the config file.
// It implements the Stringer interface that is used by the cmdopt package.
func (c ConfigFile[T]) String() string {
	if c.mountPath == "" || c.fileName == "" {
		return ""
	}

	return c.mountPath + "/" + c.fileName
}

// MountPath returns the mount path of the config file.
func (c ConfigFile[T]) MountPath() string {
	return c.mountPath
}

// FileName returns the file name of the config file.
func (c ConfigFile[T]) FileName() string {
	return c.fileName
}
