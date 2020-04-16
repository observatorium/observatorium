/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ObservatoriumSpec defines the desired state of Observatorium
type ObservatoriumSpec struct {
	// Thanos Image name
	Image string `json:"thanosImage"`
	// Thanos Image version
	Version string `json:"thanosVersion"`
	// Objest Storage Configuration
	ObjectStorageConfig ObjectStorageConfig `json:"objectStorageConfig"`
	// Hashrings describes a list of Hashrings
	Hashrings []*Hashring `json:"hashrings"`
	// Thanos CompactSpec
	Compact CompactSpec `json:"compact"`
	// Thanos Receive Controller Spec
	ReceiveController ReceiveControllerSpec `json:"thanosReceiveController"`
	// Thanos ThanosPersistentSpec
	Receivers ReceiversSpec `json:"receivers"`
	// Thanos QueryCache
	QueryCache QueryCacheSpec `json:"queryCache"`
	// Thanos StoreSpec
	Store StoreSpec `json:"store"`
	// Thanos RulerSpec
	Rule RuleSpec `json:"rule"`
	// API Gateway
	APIGateway APIGatewaySpec `json:"apiGateway"`
}

type ObjectStorageConfig struct {
	// Object Store Config Secret Name for Thanos
	Name string `json:"name"`
	// Object Store Config key for Thanos
	Key string `json:"key"`
}

type ReceiveControllerSpec struct {
	// Receive Controller image
	Image string `json:"image"`
	// Version describes the version of Thanos receive controller to use.
	Version string `json:"version,omitempty"`
}

type ReceiversSpec struct {
	// VolumeClaimTemplate
	VolumeClaimTemplate VolumeClaimTemplate `json:"volumeClaimTemplate"`
}

type StoreSpec struct {
	// VolumeClaimTemplate
	VolumeClaimTemplate VolumeClaimTemplate `json:"volumeClaimTemplate"`
	Shards              *int32              `json:"shards,omitempty"`
	// Memcached spec for Store
	Cache StoreCacheSpec `json:"cache"`
}

// StoreCacheSpec describes configuration for Store Memcached
type StoreCacheSpec struct {
	// Memcached image
	Image string `json:"image"`
	// Version of Memcached image to be deployed.
	Version string `json:"version,omitempty"`
	// Memcached Prometheus Exporter image
	ExporterImage string `json:"exporterImage"`
	// Version of Memcached Prometheus Exporter image to be deployed.
	ExporterVersion string `json:"exporterVersion,omitempty"`
	// Number of Memcached replicas.
	Replicas *int32 `json:"replicas,omitempty"`
	// Memory limit of Memcached in megabytes.
	MemoryLimitMB *int32 `json:"memoryLimitMb,omitempty"`
}

type APIGatewaySpec struct {
	// API Gateway image
	Image string `json:"image"`
	// Version describes the version of API Gateway to use.
	Version string `json:"version,omitempty"`
}

type RuleSpec struct {
	// VolumeClaimTemplate
	VolumeClaimTemplate VolumeClaimTemplate `json:"volumeClaimTemplate"`
}

type CompactSpec struct {
	// VolumeClaimTemplate
	VolumeClaimTemplate VolumeClaimTemplate `json:"volumeClaimTemplate"`
	// RetentionResolutionRaw
	RetentionResolutionRaw string `json:"retentionResolutionRaw"`
	// RetentionResolutionRaw
	RetentionResolution5m string `json:"retentionResolution5m"`
	// RetentionResolutionRaw
	RetentionResolution1h string `json:"retentionResolution1h"`
}

type VolumeClaimTemplate struct {
	Spec v1.PersistentVolumeClaimSpec `json:"spec"`
}

type QueryCacheSpec struct {
	// Thanos Query Cache image
	Image string `json:"image"`
	// Number of Query Cache replicas.
	Replicas *int32 `json:"replicas,omitempty"`
	// Version of Query Cache image to be deployed.
	Version string `json:"version,omitempty"`
}

type Hashring struct {
	// Thanos Hashring name
	Hashring string `json:"hashring"`
	// Tenants describes a lists of tenants.
	Tenants []string `json:"tenants,omitempty"`
}

// ObservatoriumStatus defines the observed state of Observatorium
type ObservatoriumStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// Observatorium is the Schema for the observatoria API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Observatorium struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ObservatoriumSpec   `json:"spec,omitempty"`
	Status ObservatoriumStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ObservatoriumList contains a list of Observatorium
type ObservatoriumList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Observatorium `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Observatorium{}, &ObservatoriumList{})
}
