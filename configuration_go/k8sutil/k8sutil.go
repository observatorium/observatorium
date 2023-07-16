package k8sutil

import (
	"fmt"

	mon "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ObjectMap represents a map of string to runtime.Objects. Usually used
// to represent a collection of manifests.
type ObjectMap map[string]runtime.Object

// Reusable K8s metadata definitions.

var ServiceMeta = metav1.TypeMeta{
	Kind:       "Service",
	APIVersion: "v1",
}

var DeploymentMeta = metav1.TypeMeta{
	Kind:       "Deployment",
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

// K8s recommended label constants.

const ComponentLabel string = "app.kubernetes.io/component"
const InstanceLabel string = "app.kubernetes.io/instance"
const NameLabel string = "app.kubernetes.io/name"
const PartOfLabel string = "app.kubernetes.io/part-of"
const VersionLabel string = "app.kubernetes.io/version"
const ManagedByLabel string = "app.kubernetes.io/managed-by"

// FlagArg returns consistent pattern flags as args for Deployment/StatefulSet containers.
// Returns empty string if flag name or value is empty. Not to be used for commands or bool args.
func FlagArg(flagName, flagValue string) string {
	if flagName == "" || flagValue == "" {
		return ""
	}

	return fmt.Sprintf("--%s=%s", flagName, flagValue)
}

// ArgList prunes any empty flags.
func ArgList(args ...string) []string {
	n := 0
	for _, x := range args {
		if x != "" {
			args[n] = x
			n++
		}
	}
	args = args[:n]
	return args
}
