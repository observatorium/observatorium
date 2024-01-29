package openshift

import (
	"fmt"
	"sort"

	"github.com/observatorium/observatorium/configuration_go/kubegen/workload"
	templatev1 "github.com/openshift/api/template/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// WrapInTemplate wraps the given runtime.Objects in an OpenShift Template, in ordered fashion.
func WrapInTemplate(objects []runtime.Object, meta metav1.ObjectMeta, params []templatev1.Parameter) *templatev1.Template {
	sort.Slice(objects, func(i, j int) bool {
		return generateObjectName(objects[i]) < generateObjectName(objects[j])
	})

	templateObjects := make([]runtime.RawExtension, len(objects))
	for i, k := range objects {
		templateObjects[i] = runtime.RawExtension{
			Object: k,
		}
	}

	return &templatev1.Template{
		TypeMeta:   workload.OpenShiftTemplateMeta,
		ObjectMeta: meta,
		Objects:    templateObjects,
		Parameters: params,
	}
}

func generateObjectName(obj runtime.Object) string {
	metaObj, ok := obj.(metav1.Object)
	if !ok {
		panic("object is not a metav1.Object")
	}

	objName := metaObj.GetName()
	if objName == "" {
		panic("object has no name")
	}

	objType := obj.GetObjectKind().GroupVersionKind().Kind

	return fmt.Sprintf("%s_%s", objName, objType)
}
