package k8sutil_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	corev1 "k8s.io/api/core/v1"
)

func TestConfigFile(t *testing.T) {
	testCases := map[string]struct {
		container         *k8sutil.Container
		option            *k8sutil.ConfigFile
		expectedContainer *k8sutil.Container
		expectedPath      string
	}{
		"empty container and no option value": {
			container: &k8sutil.Container{},
			option:    k8sutil.NewConfigFile("/etc/config", "config.yaml", "config-volume", "configmap-name"),
			expectedContainer: &k8sutil.Container{
				VolumeMounts: []corev1.VolumeMount{makeVolumeMount("config-volume", "/etc/config")},
				Volumes:      []corev1.Volume{k8sutil.NewPodVolumeFromConfigMap("config-volume", "configmap-name")},
			},
			expectedPath: "/etc/config/config.yaml",
		},
		"empty container, no option value, as secret": {
			container: &k8sutil.Container{},
			option:    k8sutil.NewConfigFile("/etc/config", "config.yaml", "config-volume", "configmap-name").AsSecret(),
			expectedContainer: &k8sutil.Container{
				VolumeMounts: []corev1.VolumeMount{makeVolumeMount("config-volume", "/etc/config")},
				Volumes:      []corev1.Volume{k8sutil.NewPodVolumeFromSecret("config-volume", "configmap-name")},
			},
			expectedPath: "/etc/config/config.yaml",
		},
		"empty container and option value as cm": {
			container: &k8sutil.Container{},
			option:    k8sutil.NewConfigFile("/etc/config", "config.yaml", "config-volume", "configmap-name").WithValue("value"),
			expectedContainer: &k8sutil.Container{
				VolumeMounts: []corev1.VolumeMount{makeVolumeMount("config-volume", "/etc/config")},
				Volumes:      []corev1.Volume{k8sutil.NewPodVolumeFromConfigMap("config-volume", "configmap-name")},
				ConfigMaps: map[string]map[string]string{
					"configmap-name": {
						"config.yaml": "value",
					},
				},
			},
			expectedPath: "/etc/config/config.yaml",
		},
		"empty container and option value as secret": {
			container: &k8sutil.Container{},
			option:    k8sutil.NewConfigFile("/etc/config", "config.yaml", "config-volume", "configmap-name").AsSecret().WithValue("value"),
			expectedContainer: &k8sutil.Container{
				VolumeMounts: []corev1.VolumeMount{makeVolumeMount("config-volume", "/etc/config")},
				Volumes:      []corev1.Volume{k8sutil.NewPodVolumeFromSecret("config-volume", "configmap-name")},
				Secrets: map[string]map[string][]byte{
					"configmap-name": {
						"config.yaml": []byte("value"),
					},
				},
			},
			expectedPath: "/etc/config/config.yaml",
		},
		"already existing volume": {
			container: &k8sutil.Container{
				Volumes: []corev1.Volume{k8sutil.NewPodVolumeFromConfigMap("config-volume", "configmap-name")},
			},
			option: k8sutil.NewConfigFile("/etc/config", "config.yaml", "config-volume-other", "configmap-name"),
			expectedContainer: &k8sutil.Container{
				Volumes: []corev1.Volume{k8sutil.NewPodVolumeFromConfigMap("config-volume", "configmap-name")},
				VolumeMounts: []corev1.VolumeMount{
					makeVolumeMount("config-volume", "/etc/config"),
				},
			},
			expectedPath: "/etc/config/config.yaml",
		},
		"already existing volume, as secret": {
			container: &k8sutil.Container{
				Volumes: []corev1.Volume{k8sutil.NewPodVolumeFromSecret("config-volume", "configmap-name")},
			},
			option: k8sutil.NewConfigFile("/etc/config", "config.yaml", "config-volume-other", "configmap-name").AsSecret(),
			expectedContainer: &k8sutil.Container{
				Volumes: []corev1.Volume{k8sutil.NewPodVolumeFromSecret("config-volume", "configmap-name")},
				VolumeMounts: []corev1.VolumeMount{
					makeVolumeMount("config-volume", "/etc/config"),
				},
			},
			expectedPath: "/etc/config/config.yaml",
		},
		"already existing volume and mount": { // check that the mount path is updated
			container: &k8sutil.Container{
				Volumes:      []corev1.Volume{k8sutil.NewPodVolumeFromConfigMap("config-volume", "configmap-name")},
				VolumeMounts: []corev1.VolumeMount{makeVolumeMount("config-volume", "/etc/config")},
			},
			option: k8sutil.NewConfigFile("/etc/config-other-path", "other-config.yaml", "config-volume", "configmap-name"),
			expectedContainer: &k8sutil.Container{
				VolumeMounts: []corev1.VolumeMount{makeVolumeMount("config-volume", "/etc/config")},
				Volumes:      []corev1.Volume{k8sutil.NewPodVolumeFromConfigMap("config-volume", "configmap-name")},
			},
			expectedPath: "/etc/config/other-config.yaml",
		},
		"already existing volume and mount as secret": { // check that the mount path is updated
			container: &k8sutil.Container{
				Volumes:      []corev1.Volume{k8sutil.NewPodVolumeFromSecret("config-volume", "configmap-name")},
				VolumeMounts: []corev1.VolumeMount{makeVolumeMount("config-volume", "/etc/config")},
			},
			option: k8sutil.NewConfigFile("/etc/config-other-path", "other-config.yaml", "config-volume", "configmap-name").AsSecret(),
			expectedContainer: &k8sutil.Container{
				VolumeMounts: []corev1.VolumeMount{makeVolumeMount("config-volume", "/etc/config")},
				Volumes:      []corev1.Volume{k8sutil.NewPodVolumeFromSecret("config-volume", "configmap-name")},
			},
			expectedPath: "/etc/config/other-config.yaml",
		},
		"with existing config map": {
			container: &k8sutil.Container{},
			option:    k8sutil.NewConfigFile("/etc/config", "config.yaml", "config-volume", "configmap-name").WithExistingResource("myconfig", "myconfig.yaml"),
			expectedContainer: &k8sutil.Container{
				VolumeMounts: []corev1.VolumeMount{makeVolumeMount("config-volume", "/etc/config")},
				Volumes:      []corev1.Volume{k8sutil.NewPodVolumeFromConfigMap("config-volume", "myconfig")},
			},
			expectedPath: "/etc/config/myconfig.yaml",
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			testCase.option.AddToContainer(testCase.container)
			compareContainers(testCase.container, testCase.expectedContainer, t)
		})
	}
}

func compareCMOrSecret[T string | []byte](have, expect map[string]map[string]T, t *testing.T) {
	if len(have) != len(expect) {
		t.Fatalf("expected %d config maps, got %d", len(expect), len(have))
	}

	for haveName, haveValue := range have {
		expectValue, ok := expect[haveName]
		if !ok {
			t.Fatalf("expected config map %s to exist", haveName)
		}
		for haveKey, haveCMValue := range haveValue {
			expectValue, ok := expectValue[haveKey]
			if !ok {
				t.Fatalf("expected config map %s to have key %s", haveName, haveKey)
			}
			if !reflect.DeepEqual(haveCMValue, expectValue) {
				t.Fatalf("expected config map %s to have value %v, got %v", haveName, expectValue, haveCMValue)
			}
		}

	}
}

func compareContainers(have, expect *k8sutil.Container, t *testing.T) {
	if len(have.VolumeMounts) != len(expect.VolumeMounts) {
		t.Fatalf("expected %d volume mounts, got %d", len(expect.VolumeMounts), len(have.VolumeMounts))
	}

	for i, haveMount := range have.VolumeMounts {
		expectMount := expect.VolumeMounts[i]
		if haveMount.Name != expectMount.Name {
			t.Fatalf("expected volume mount name to be %s, got %s", expectMount.Name, haveMount.Name)
		}
		if haveMount.MountPath != expectMount.MountPath {
			t.Fatalf("expected volume mount path to be %s, got %s", expectMount.MountPath, haveMount.MountPath)
		}
	}

	if len(have.Volumes) != len(expect.Volumes) {
		fmt.Println(have.Volumes)
		t.Fatalf("expected %d volumes, got %d", len(expect.Volumes), len(have.Volumes))
	}

	for i, haveVolume := range have.Volumes {
		expectVolume := expect.Volumes[i]
		if haveVolume.Name != expectVolume.Name {
			t.Fatalf("expected volume name to be %s, got %s", expectVolume.Name, haveVolume.Name)
		}
		if !reflect.DeepEqual(haveVolume.VolumeSource, expectVolume.VolumeSource) {
			t.Fatalf("expected volume source to be %v, got %v", expectVolume.VolumeSource, haveVolume.VolumeSource)
		}
	}

	if len(have.Secrets) != len(expect.Secrets) {
		t.Fatalf("expected %d secrets, got %d", len(expect.Secrets), len(have.Secrets))
	}

	compareCMOrSecret(have.Secrets, expect.Secrets, t)

	if len(have.ConfigMaps) != len(expect.ConfigMaps) {
		t.Fatalf("expected %d config maps, got %d", len(expect.ConfigMaps), len(have.ConfigMaps))
	}

	compareCMOrSecret(have.ConfigMaps, expect.ConfigMaps, t)
}

func makeVolumeMount(name, path string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      name,
		MountPath: path,
	}
}
