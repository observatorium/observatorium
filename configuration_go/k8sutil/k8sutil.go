package k8sutil

import (
	"fmt"

	mon "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ObjectMap represents a map of string to runtime.Objects. Usually used
// to represent a collection of manifests.
type ObjectMap map[string]runtime.Object

func (o ObjectMap) Add(obj runtime.Object) {
	metaObj, ok := obj.(metav1.Object)
	if !ok {
		panic(fmt.Sprintf("object %v has no name", obj))
	}

	objName := metaObj.GetName()
	if objName == "" {
		panic(fmt.Sprintf("object %v has no name", obj))
	}

	objType := obj.GetObjectKind().GroupVersionKind().Kind

	if _, ok := o[objName]; ok {
		panic(fmt.Sprintf("object %s/%s already exists", objType, objName))
	}

	o[o.makeKey(objType, objName)] = obj
}

func (o ObjectMap) makeKey(objType, objName string) string {
	return fmt.Sprintf("%s/%s", objName, objType)
}

func (o ObjectMap) AddAll(objs []runtime.Object) {
	for _, obj := range objs {
		o.Add(obj)
	}
}

// GetObject returns the object of type T from the given map of kubernetes objects.
// When specifying a name, it will return the object with the given name.
// This helper can be used for doing post processing on the objects.
func GetObject[T metav1.Object](manifests ObjectMap, name string) T {
	var ret T
	found := false

	for _, obj := range manifests {
		if service, ok := obj.(T); ok {
			if name != "" && service.GetName() != name {
				continue
			}

			// Check if we already found an object of this type. If so, panic.
			if found {
				panic(fmt.Sprintf("found multiple objects of type %T", *new(T)))
			}

			ret = service
			found = true
		}
	}

	if !found {
		panic(fmt.Sprintf("could not find object of type %T", *new(T)))
	}

	return ret
}

// Reusable K8s metadata definitions.

var ServiceMeta = metav1.TypeMeta{
	Kind:       "Service",
	APIVersion: "v1",
}

var DeploymentMeta = metav1.TypeMeta{
	Kind:       "Deployment",
	APIVersion: "apps/v1",
}

var StatefulSetMeta = metav1.TypeMeta{
	Kind:       "StatefulSet",
	APIVersion: "apps/v1",
}

var SecretMeta = metav1.TypeMeta{
	Kind:       "Secret",
	APIVersion: "v1",
}

var ConfigMapMeta = metav1.TypeMeta{
	Kind:       "ConfigMap",
	APIVersion: "v1",
}

var ServiceAccountMeta = metav1.TypeMeta{
	Kind:       "ServiceAccount",
	APIVersion: "v1",
}

var ServiceMonitorMeta = metav1.TypeMeta{
	Kind:       monv1.ServiceMonitorsKind,
	APIVersion: fmt.Sprintf("%s/%s", mon.GroupName, monv1.Version),
}

var OpenShiftTemplateMeta = metav1.TypeMeta{
	Kind:       "Template",
	APIVersion: "template.openshift.io/v1",
}

var RoleMeta = metav1.TypeMeta{
	Kind:       "Role",
	APIVersion: rbacv1.SchemeGroupVersion.String(),
}

var RoleBindingMeta = metav1.TypeMeta{
	Kind:       "RoleBinding",
	APIVersion: rbacv1.SchemeGroupVersion.String(),
}

// K8s recommended label constants.

const ComponentLabel string = "app.kubernetes.io/component"
const InstanceLabel string = "app.kubernetes.io/instance"
const NameLabel string = "app.kubernetes.io/name"
const PartOfLabel string = "app.kubernetes.io/part-of"
const VersionLabel string = "app.kubernetes.io/version"
const ManagedByLabel string = "app.kubernetes.io/managed-by"
const HostnameLabel string = "kubernetes.io/hostname"
const OsLabel string = "kubernetes.io/os"
const LinuxOs string = "linux"

// // FlagArg returns consistent pattern flags as args for Deployment/StatefulSet containers.
// // Returns empty string if flag name or value is empty or if flag value is a zero/default value.
// // Not to be used for commands or bool args.
// func FlagArg(flagName, flagValue string) string {
// 	if flagName == "" || flagValue == "" || flagValue == "0" || flagValue == "0s" {
// 		return ""
// 	}

// 	return fmt.Sprintf("--%s=%s", flagName, flagValue)
// }

// // BoolFlagArg returns consistent pattern bool flags as args for Deployment/StatefulSet containers.
// // Returns empty string if flag name is empty or value is false.
// func BoolFlagArg(flagName string, flagValue bool) string {
// 	if flagName == "" || !flagValue {
// 		return ""
// 	}

// 	return fmt.Sprintf("--%s", flagName)
// }

// // RepeatableFloatFlagArg returns consistent pattern repeatable flags as args for Deployment/StatefulSet containers.
// func RepeatableFloatFlagArg(flagName string, flagValues []float64) []string {
// 	if flagName == "" || len(flagValues) == 0 {
// 		return []string{}
// 	}

// 	result := []string{}
// 	for _, v := range flagValues {
// 		result = append(result, fmt.Sprintf("--%s=%f", flagName, v))
// 	}

// 	return result
// }

// // RepeatableFlagArg returns consistent pattern repeatable flags as args for Deployment/StatefulSet containers.
// func RepeatableFlagArg(flagName string, flagValues []string) []string {
// 	if flagName == "" || len(flagValues) == 0 {
// 		return []string{}
// 	}

// 	result := []string{}
// 	for _, v := range flagValues {
// 		result = append(result, fmt.Sprintf("--%s=%s", flagName, v))
// 	}

// 	return result
// }

// // RepeatableFlagArg returns consistent pattern repeatable flags as args for Deployment/StatefulSet containers.
// func RepeatableLabelFlagArg(flagName string, flagValues labels.Labels) []string {
// 	if flagName == "" || len(flagValues) == 0 {
// 		return []string{}
// 	}

// 	result := []string{}

// 	fs := flagValues.String()
// 	fs = fs[1 : len(fs)-2]
// 	ls := strings.Split(fs, ", ")

// 	for _, v := range ls {
// 		result = append(result, fmt.Sprintf("--%s=%s", flagName, v))
// 	}

// 	return result
// }

// // ArgList prunes any empty flags.
// func ArgList(args []string) []string {
// 	n := 0
// 	for _, x := range args {
// 		if x != "" {
// 			args[n] = x
// 			n++
// 		}
// 	}
// 	args = args[:n]
// 	return args
// }
