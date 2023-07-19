package k8sutil

import (
	"fmt"
	"strings"

	mon "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/model/labels"
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
// Returns empty string if flag name or value is empty or if flag value is a zero/default value.
// Not to be used for commands or bool args.
func FlagArg(flagName, flagValue string) string {
	if flagName == "" || flagValue == "" || flagValue == "0" || flagValue == "0s" {
		return ""
	}

	return fmt.Sprintf("--%s=%s", flagName, flagValue)
}

// BoolFlagArg returns consistent pattern bool flags as args for Deployment/StatefulSet containers.
// Returns empty string if flag name is empty or value is false.
func BoolFlagArg(flagName string, flagValue bool) string {
	if flagName == "" || !flagValue {
		return ""
	}

	return fmt.Sprintf("--%s", flagName)
}

// RepeatableFloatFlagArg returns consistent pattern repeatable flags as args for Deployment/StatefulSet containers.
func RepeatableFloatFlagArg(flagName string, flagValues []float64) []string {
	if flagName == "" || len(flagValues) == 0 {
		return []string{}
	}

	result := []string{}
	for _, v := range flagValues {
		result = append(result, fmt.Sprintf("--%s=%f", flagName, v))
	}

	return result
}

// RepeatableFlagArg returns consistent pattern repeatable flags as args for Deployment/StatefulSet containers.
func RepeatableFlagArg(flagName string, flagValues []string) []string {
	if flagName == "" || len(flagValues) == 0 {
		return []string{}
	}

	result := []string{}
	for _, v := range flagValues {
		result = append(result, fmt.Sprintf("--%s=%s", flagName, v))
	}

	return result
}

// RepeatableFlagArg returns consistent pattern repeatable flags as args for Deployment/StatefulSet containers.
func RepeatableLabelFlagArg(flagName string, flagValues labels.Labels) []string {
	if flagName == "" || len(flagValues) == 0 {
		return []string{}
	}

	result := []string{}

	fs := flagValues.String()
	fs = fs[1 : len(fs)-2]
	ls := strings.Split(fs, ", ")

	for _, v := range ls {
		result = append(result, fmt.Sprintf("--%s=%s", flagName, v))
	}

	return result
}

// ArgList prunes any empty flags.
func ArgList(args []string) []string {
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
