package workload

import (
	"fmt"

	mon "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
