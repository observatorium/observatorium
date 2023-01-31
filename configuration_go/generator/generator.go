package generator

import (
	"github.com/bwplotka/mimic"
	"github.com/bwplotka/mimic/encoding"
	"github.com/observatorium/observatorium/configuration_go/k8sutil"
)

// GenerateWithMimic renders given ObjectMap to YAML with github.com/bwplotka/mimic.
func GenerateWithMimic(g *mimic.Generator, objMap k8sutil.ObjectMap, parts ...string) {
	for name, manifest := range objMap {
		g.With(parts...).Add(name+".yaml", encoding.GhodssYAML(manifest))
	}
}
