package k8sutil

import (
	"fmt"
	"path/filepath"
	"reflect"

	corev1 "k8s.io/api/core/v1"
)

// ConfigFileWithType represents a configuration file that must be consumed by a container.
// It is similar to ConfigFile but it allows to specify the type of the value.
type ConfigFileWithType[T fmt.Stringer] struct {
	ConfigFile
}

// NewConfigFileWithType creates a new ConfigFileWithType.
func NewConfigFileWithType[T fmt.Stringer](mountPath, key, volumeName, resourceName string) *ConfigFileWithType[T] {
	ret := &ConfigFileWithType[T]{
		ConfigFile: ConfigFile{
			mountPath:    mountPath,
			volumeName:   volumeName,
			key:          key,
			resourceName: resourceName,
		},
	}

	return ret
}

// WithValue creates a new resource (ConfigMap or Secret) with the given value.
func (c *ConfigFileWithType[T]) WithValue(value T) *ConfigFileWithType[T] {
	c.value = value.String()
	return c
}

// ConfigFile represents a configuration file that must be consumed by a container.
// It encapsulates the data needed to mount the file as a configMap or secret in a container.
// Usage:
//
//	configFile := k8sutil.NewConfigFile("/etc/config", "config.yaml", "config-file", "my-config")
//	configFile.WithValue("value").AddToContainer(container)
type ConfigFile struct {
	resourceName string
	mountPath    string
	volumeName   string
	key          string
	value        string
	isSecret     bool
}

// NewConfigFile creates a new ConfigFile where the date type is a string.
func NewConfigFile(mountPath, key, volumeName, cmName string) *ConfigFile {
	ret := &ConfigFile{
		mountPath:    mountPath,
		volumeName:   volumeName,
		key:          key,
		resourceName: cmName,
	}

	return ret
}

// AsSecret specifies that the resource must be a secret instead of a configMap.
func (c *ConfigFile) AsSecret() *ConfigFile {
	c.isSecret = true
	return c
}

// WithExistingResource specifies the name of the resource (ConfigMap or Secret) and the key to use.
// It is used when the resource already exists.
func (c ConfigFile) WithExistingResource(name, key string) *ConfigFile {
	c.resourceName = name
	c.key = key
	return &c
}

// WithValue creates a new resource (ConfigMap or Secret) with a string value.
func (c ConfigFile) WithValue(value string) *ConfigFile {
	c.value = value
	return &c
}

// WithResourceName specifies the name of the resource (ConfigMap or Secret) to create.
// It can be used to override the default resource name when using With*Value().
func (c ConfigFile) WithResourceName(resourceName string) *ConfigFile {
	c.resourceName = resourceName
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
// It configures volumes, volume mounts and resources (ConfigMap or Secret) for the container.
// It includes some logic to avoid creating duplicate resources, volumes and volume mounts.
func (c *ConfigFile) AddToContainer(container *Container) {
	if c.resourceName == "" {
		panic("resource name is empty")
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

	c.addResourceToContainer(container)
	c.addVolumeToContainer(container)
	c.addVolumeMoutToContainer(container)
}

func (c *ConfigFile) addResourceToContainer(container *Container) {
	// If resource must be created, add it to the container
	if c.value == "" {
		return
	}

	if c.isSecret {
		if container.Secrets == nil {
			container.Secrets = make(map[string]map[string][]byte)
		}

		newSecret := map[string][]byte{
			c.key: []byte(c.value),
		}

		// check if secret already exists
		if val, ok := container.Secrets[c.resourceName]; ok {
			// Check if content is the same
			if reflect.DeepEqual(val, newSecret) {
				return
			}

			panic(fmt.Sprintf("secret %q already exists", c.resourceName))
		}

		container.Secrets[c.resourceName] = newSecret
	} else {
		if container.ConfigMaps == nil {
			container.ConfigMaps = make(map[string]map[string]string)
		}

		newConfigMap := map[string]string{
			c.key: c.value,
		}

		// check if configmap already exists
		if val, ok := container.ConfigMaps[c.resourceName]; ok {
			// Check if content is the same
			if reflect.DeepEqual(val, newConfigMap) {
				return
			}

			panic(fmt.Sprintf("configmap %q already exists", c.resourceName))
		}

		container.ConfigMaps[c.resourceName] = newConfigMap
	}
}

func (c *ConfigFile) addVolumeToContainer(container *Container) {
	// Check if a volume with this resource name already exists
	// var existingVolume *corev1.Volume
	for _, vol := range container.Volumes {
		if c.isSecret && vol.VolumeSource.Secret != nil && vol.VolumeSource.Secret.SecretName == c.resourceName {
			// Update volume name
			c.volumeName = vol.Name
			// existingVolume = &vol
			return
		}

		if !c.isSecret && vol.VolumeSource.ConfigMap != nil && vol.VolumeSource.ConfigMap.Name == c.resourceName {
			// Update volume name
			c.volumeName = vol.Name
			// existingVolume = &vol
			return
		}
	}

	// Check that the volume name is not already used
	for _, vol := range container.Volumes {
		if vol.Name == c.volumeName {
			panic(fmt.Sprintf("volume name %q is already used", c.volumeName))
		}
	}

	// Add the volume to the container
	if c.isSecret {
		container.Volumes = append(container.Volumes, NewPodVolumeFromSecret(c.volumeName, c.resourceName))
	} else {
		container.Volumes = append(container.Volumes, NewPodVolumeFromConfigMap(c.volumeName, c.resourceName))
	}
}

func (c *ConfigFile) addVolumeMoutToContainer(container *Container) {
	// Check if the volume is already mounted and update mount path
	for _, mount := range container.VolumeMounts {
		if mount.Name == c.volumeName {
			c.mountPath = mount.MountPath
			return
		}
	}

	// Check if mount path is already used
	for _, mount := range container.VolumeMounts {
		if mount.MountPath == c.mountPath {
			panic(fmt.Sprintf("mount path %q is already used", c.mountPath))
		}
	}

	// Add the volume mount to the container
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      c.volumeName,
		MountPath: c.mountPath,
		ReadOnly:  true,
	})
}
