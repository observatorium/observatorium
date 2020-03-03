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
	// Thanos QuerierCache
	QueryCache QuerierCacheSpec `json:"queryCache"`
	// Thanos StoreSpec
	Store StoreSpec `json:"store"`
	// Thanos RulerSpec
	Rule RuleSpec `json:"rule"`
	// API Gateway
	ApiGateway ApiGatewaySpec `json:"apiGateway"`
}

type ObjectStorageConfig struct {
	// Object Store Config Secret Name for Thanos
	Name string `json:"name"`
	// Object Store Config key for Thanos
	Key string `json:"key"`
}

type ReceiveControllerSpec struct {
	Image string `json:"image"`
	// Tag describes the tag of Thanos receive controller to use.
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
}

type ApiGatewaySpec struct {
	Image string `json:"image"`
	// Tag describes the tag of Thanos receive controller to use.
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
	VolumeClaimTemplateSpec v1.PersistentVolumeClaimSpec `json:"spec"`
}

type QuerierCacheSpec struct {
	// Thanos Querier Cache  name
	Image string `json:"image"`
	// Number of Querier Cache replicas.
	Replicas *int32 `json:"replicas,omitempty"`
	// Version of  Querier Cache image to be deployed.
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
