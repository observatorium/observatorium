package openshift

import (
	"sort"

	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	templatev1 "github.com/openshift/api/template/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// WrapInTemplate wraps the given runtime.Objects in an OpenShift Template, in ordered fashion.
func WrapInTemplate(name string, objMap k8sutil.ObjectMap, meta metav1.ObjectMeta, params []templatev1.Parameter) k8sutil.ObjectMap {
	keys := sortedKeys(objMap)

	objects := make([]runtime.RawExtension, len(objMap))
	for i, k := range keys {
		objects[i] = runtime.RawExtension{
			Object: objMap[k],
		}
	}

	return k8sutil.ObjectMap{
		name: &templatev1.Template{
			TypeMeta:   k8sutil.OpenShiftTemplateMeta,
			ObjectMeta: meta,
			Objects:    objects,
			Parameters: params,
		},
	}
}

// sortedKeys helps in traversing map in ordered fashion, for consistent diffs.
func sortedKeys(objMap k8sutil.ObjectMap) []string {
	keys := []string{}
	for k := range objMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}
