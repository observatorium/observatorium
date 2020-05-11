package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	yaml "github.com/ghodss/yaml"
	csvv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// CSVClusterPermissions is the cluster permissions part of a CSV
type CSVClusterPermissions struct {
	ServiceAccountName string              `json:"serviceAccountName"`
	Rules              []rbacv1.PolicyRule `json:"rules"`
}

// CSVPermissions is the permissions part of a CSV
type CSVPermissions struct {
	ServiceAccountName string              `json:"serviceAccountName"`
	Rules              []rbacv1.PolicyRule `json:"rules"`
}

// CSVDeployments describes the deployments for a CSV
type CSVDeployments struct {
	Name string                `json:"name"`
	Spec appsv1.DeploymentSpec `json:"spec,omitempty"`
}

// CSVStrategySpec describes the installation strategy of a CSV
type CSVStrategySpec struct {
	ClusterPermissions []CSVClusterPermissions `json:"clusterPermissions"`
	Permissions        []CSVPermissions        `json:"permissions"`
	Deployments        []CSVDeployments        `json:"deployments"`
}

// UnmarshalCSV decodes a YAML file, by path, and returns a CSV
func UnmarshalCSV(filePath string) *csvv1.ClusterServiceVersion {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	csvStruct := &csvv1.ClusterServiceVersion{}
	err = yaml.Unmarshal(bytes, csvStruct)
	if err != nil {
		panic(err)
	}

	return csvStruct
}

// UnmarshalStrategySpec decodes the StrategySpec object inside a CSV and returns it
func UnmarshalStrategySpec(csv *csvv1.ClusterServiceVersion) *CSVStrategySpec {
	templateStrategySpec := &CSVStrategySpec{}
	err := json.Unmarshal(csv.Spec.InstallStrategy.StrategySpecRaw, templateStrategySpec)
	if err != nil {
		panic(err)
	}

	return templateStrategySpec
}

// MarshallObject mashals an object, usually a CSV into YAML
func MarshallObject(obj interface{}, writer io.Writer) error {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	var r unstructured.Unstructured
	if err := json.Unmarshal(jsonBytes, &r.Object); err != nil {
		return err
	}

	// remove status and metadata.creationTimestamp
	unstructured.RemoveNestedField(r.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(r.Object, "template", "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(r.Object, "spec", "template", "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(r.Object, "status")

	deployments, exists, err := unstructured.NestedSlice(r.Object, "spec", "install", "spec", "deployments")
	if exists {
		for _, obj := range deployments {
			deployment := obj.(map[string]interface{})
			unstructured.RemoveNestedField(deployment, "metadata", "creationTimestamp")
			unstructured.RemoveNestedField(deployment, "spec", "template", "metadata", "creationTimestamp")
			unstructured.RemoveNestedField(deployment, "status")
		}
		unstructured.SetNestedSlice(r.Object, deployments, "spec", "install", "spec", "deployments")
	}

	jsonBytes, err = json.Marshal(r.Object)
	if err != nil {
		return err
	}

	yamlBytes, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		return err
	}

	// fix double quoted strings by removing unneeded single quotes...
	s := string(yamlBytes)
	s = strings.Replace(s, " '\"", " \"", -1)
	s = strings.Replace(s, "\"'\n", "\"\n", -1)

	yamlBytes = []byte(s)

	_, err = writer.Write([]byte("---\n"))
	if err != nil {
		return err
	}

	_, err = writer.Write(yamlBytes)
	if err != nil {
		return err
	}

	return nil
}

