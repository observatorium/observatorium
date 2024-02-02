package kubeyaml

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bwplotka/mimic"
	"github.com/bwplotka/mimic/encoding"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

type FilenameGenerator func(obj runtime.Object) string

type MimicEncoder func(...interface{}) encoding.Encoder

type ObjectSerializer func(interface{}) ([]byte, error)

// GenerateWithMimic renders given objects to YAML with github.com/bwplotka/mimic.
func GenerateWithMimic(g *mimic.Generator, objects []runtime.Object, parts ...string) {
	GenerateWithMimicCustomFilename(g, objects, KubeObjectNameAndKind, parts...)
}

// GenerateWithMimic and specify filename generator.
func GenerateWithMimicCustomFilename(g *mimic.Generator, objects []runtime.Object, fg FilenameGenerator, parts ...string) {
	for _, obj := range objects {
		g.With(parts...).Add(fg(obj)+".yaml", encoding.GhodssYAML(obj))
	}
}

func KubeObjectNameAndKind(obj runtime.Object) string {
	metaObj, ok := obj.(metav1.Object)
	if !ok {
		panic(fmt.Sprintf("object %v has no name", obj))
	}

	objName := metaObj.GetName()
	if objName == "" {
		panic(fmt.Sprintf("object %v has no name", obj))
	}

	objType := obj.GetObjectKind().GroupVersionKind().Kind

	return fmt.Sprintf("%s_%s", objName, objType)
}

// WriteObjectsInDir writes given objects to the given directory.
func WriteObjectsInDir(objects []runtime.Object, dir string) {
	WriteObjectsInDirWithSerializer(objects, dir, yaml.Marshal)
}

// WriteObjectsInDirWithSerializer writes given objects to the given directory.
func WriteObjectsInDirWithSerializer(objects []runtime.Object, dir string, enc ObjectSerializer) {
	for _, obj := range objects {
		name := KubeObjectNameAndKind(obj)
		path := filepath.Join(dir, name) + ".yaml"
		data, err := yaml.Marshal(obj)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal manifest %s: %v", name, err))
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(fmt.Sprintf("failed to create directory %s: %v", dir, err))
		}

		if err := os.WriteFile(path, data, 0644); err != nil {
			panic(fmt.Sprintf("failed to write manifest %s: %v", name, err))
		}
	}
}
