package containeropts

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/observatorium/observatorium/configuration_go/kubegen/helpers"
	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	corev1 "k8s.io/api/core/v1"
)

type ContainerUpdater interface {
	Update(*workload.Container)
}

// FileInVolume is a configuration file that must be consumed by the container.
// It encapsulates the data needed to mount the file as a volume in a container.
// It is used when the config file is generated by a sidecar container or an init container.
// It mounts the volume in the container and provides the path to the config file in the String() method.
type FileInVolume struct {
	volumeName string
	mountPath  string
	filePath   string
}

// NewFileInVolume creates a new VolumeConfigFile.
func NewFileInVolume(volumeName, mountPath, filepath string) *FileInVolume {
	return &FileInVolume{
		volumeName: volumeName,
		mountPath:  mountPath,
		filePath:   filepath,
	}
}

// String returns the path to the config file.
func (c *FileInVolume) String() string {
	return filepath.Join(c.mountPath, c.filePath)
}

// Update mount the required volume to the container.
func (c *FileInVolume) Update(container *workload.Container) {
	addVolumeMountToContainer(container, c.volumeName, c.mountPath)
}

// ConfigResourceAsFile represents a configuration file that must be consumed by a container.
// It encapsulates the data needed to mount the file as a configMap or secret in a container.
type ConfigResourceAsFile struct {
	resourceName string
	mountPath    string
	volumeName   string
	key          string
	value        string
	isSecret     bool
}

// NewConfigResourceAsFile creates a new ConfigFile.
func NewConfigResourceAsFile(mountPath, key, volumeName, resourceName string) *ConfigResourceAsFile {
	return &ConfigResourceAsFile{
		mountPath:    mountPath,
		volumeName:   volumeName,
		key:          key,
		resourceName: resourceName,
	}
}

// WithValue sets the resource's (ConfigMap or Secret) value.
func (c *ConfigResourceAsFile) WithValue(value string) *ConfigResourceAsFile {
	c.value = value

	return c
}

// AsSecret specifies that the resource must be a secret instead of a configMap (the default).
func (c *ConfigResourceAsFile) AsSecret() *ConfigResourceAsFile {
	c.isSecret = true
	return c
}

// AsConfigMap specifies that the resource must be a configMap instead of a secret.
// It is used when the config file is set as a secret by default.
func (c *ConfigResourceAsFile) AsConfigMap() *ConfigResourceAsFile {
	c.isSecret = false
	return c
}

// WithExistingResource specifies the name of the resource (ConfigMap or Secret) and the key to use.
// It is used when the resource already exists.
func (c *ConfigResourceAsFile) WithExistingResource(name, key string) *ConfigResourceAsFile {
	c.resourceName = name
	c.key = key
	return c
}

// WithResourceName specifies the name of the resource (ConfigMap or Secret) to create.
// It can be used to override the default resource name when using WithValue().
func (c *ConfigResourceAsFile) WithResourceName(resourceName string) *ConfigResourceAsFile {
	c.resourceName = resourceName
	return c
}

// String returns the path to the config file.
// It implements the Stringer interface that is used by the cmdopt package.
func (c *ConfigResourceAsFile) String() string {
	if c.mountPath == "" || c.key == "" {
		return ""
	}

	return filepath.Join(c.mountPath, c.key)
}

// Update adds the config file to the container.
// It configures volumes, volume mounts and resources (ConfigMap or Secret) for the container.
// It includes some logic to avoid creating duplicate resources, volumes and volume mounts.
func (c *ConfigResourceAsFile) Update(container *workload.Container) {
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
	addVolumeMountToContainer(container, c.volumeName, c.mountPath)
}

func (c *ConfigResourceAsFile) addResourceToContainer(container *workload.Container) {
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

func (c *ConfigResourceAsFile) addVolumeToContainer(container *workload.Container) {
	// Check if a volume with this resource name already exists
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
		container.Volumes = append(container.Volumes, helpers.NewPodVolumeFromSecret(c.volumeName, c.resourceName))
	} else {
		container.Volumes = append(container.Volumes, helpers.NewPodVolumeFromConfigMap(c.volumeName, c.resourceName))
	}
}

func addVolumeMountToContainer(container *workload.Container, volumeName, mountPath string) {
	// Check if the volume is already mounted
	for _, mount := range container.VolumeMounts {
		if mount.Name == volumeName {
			return
		}
	}

	// Check if mount path is already used
	for _, mount := range container.VolumeMounts {
		if strings.HasPrefix(mountPath, mount.MountPath) {
			panic(fmt.Sprintf("mount path %q is already used", mountPath))
		}
	}

	// Add the volume mount to the container
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      volumeName,
		MountPath: mountPath,
		ReadOnly:  true,
	})
}
