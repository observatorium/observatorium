package k8sutil

import (
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
)

type ConfigFileWithType[T fmt.Stringer] struct {
	ConfigFile
}

func NewConfigFileWithType[T fmt.Stringer](mountPath, key, volumeName, cmName string) *ConfigFileWithType[T] {
	ret := &ConfigFileWithType[T]{
		ConfigFile: ConfigFile{
			mountPath:     mountPath,
			volumeName:    volumeName,
			key:           key,
			configMapName: cmName,
		},
	}

	return ret
}

func (c *ConfigFileWithType[T]) WithValue(value T) *ConfigFileWithType[T] {
	c.value = value.String()
	return c
}

func (c *ConfigFileWithType[T]) WithStrValue(value string) *ConfigFileWithType[T] {
	c.value = value
	return c
}

// ConfigFile is a struct that represents a config file.
// It includes data needed to mount the file as a configMap or secret in a container.
type ConfigFile struct {
	configMapName string
	mountPath     string
	volumeName    string
	key           string
	value         string
}

// NewConfigFile creates a new ConfigFile.
func NewConfigFile(mountPath, key, volumeName, cmName string) *ConfigFile {
	ret := &ConfigFile{
		mountPath:     mountPath,
		volumeName:    volumeName,
		key:           key,
		configMapName: cmName,
	}

	return ret
}

// WithExistingCM creates a new ConfigFile with an existing configmap.
func (c ConfigFile) WithExistingCM(cmName, key string) *ConfigFile {
	c.configMapName = cmName
	c.key = key
	return &c
}

// WithStrValue creates a new ConfigFile with a string value.
func (c ConfigFile) WithValue(value string) *ConfigFile {
	// c.configMapName = cmName
	c.value = value
	return &c
}

// String returns the path to the config file.
// It implements the Stringer interface that is used by the cmdopt package.
func (c ConfigFile) String() string {
	if c.mountPath == "" || c.key == "" {
		return ""
	}

	return filepath.Join(c.mountPath, c.key)
}

// AddToContainer adds the config file to the container.
func (c *ConfigFile) AddToContainer(container *Container) {
	if c.configMapName == "" {
		panic("configmap name is empty")
	}

	if c.mountPath == "" {
		panic("mount path is empty")
	}

	if c.volumeName == "" {
		panic("volume name is empty")
	}

	if c.key == "" {
		panic("key is empty")
	}

	if c.value != "" {
		container.ConfigMaps[c.configMapName] = map[string]string{
			c.key: c.value,
		}
	}

	// Check if configmap is already mounted
	for _, vol := range container.Volumes {
		if vol.ConfigMap == nil || vol.ConfigMap.Name != c.configMapName {
			continue
		}
		for _, mount := range container.VolumeMounts {
			if mount.Name == vol.Name {
				c.mountPath = mount.MountPath
				return
			}
		}
	}

	// Check if mount path is already used
	for _, mount := range container.VolumeMounts {
		if mount.MountPath == c.mountPath {
			panic(fmt.Sprintf("mount path %q is already used", c.mountPath))
		}
	}

	container.Volumes = append(container.Volumes, NewPodVolumeFromConfigMap(c.volumeName, c.configMapName))
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      c.volumeName,
		MountPath: c.mountPath,
	})
}
