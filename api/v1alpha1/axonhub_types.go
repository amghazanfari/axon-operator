/*
Copyright 2026 amghazanfari.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AxonHubSpec defines the desired state of AxonHub.
type AxonHubSpec struct {
	// Image is the AxonHub (octopus) container image.
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Replicas is the number of AxonHub instances to run.
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	Replicas *int32 `json:"replicas,omitempty"`

	// Port is the port the AxonHub container listens on.
	// +kubebuilder:default=8090
	Port int32 `json:"port,omitempty"`

	// Postgres defines the database configuration for this AxonHub instance.
	// +kubebuilder:validation:Required
	Postgres PostgresConfig `json:"postgres"`

	// Resources defines compute resources for the AxonHub container.
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Env allows passing extra environment variables to the AxonHub container.
	Env []corev1.EnvVar `json:"env,omitempty"`
}

// PostgresConfig defines the database configuration for AxonHub.
// When Enabled is true, the operator creates and manages a dedicated Postgres
// instance for this AxonHub CR. When Enabled is false, the operator expects
// an External configuration pointing to an existing database.
type PostgresConfig struct {
	// Enabled specifies whether the operator should create a dedicated Postgres
	// instance. If false, External connection details must be provided.
	// +kubebuilder:default=true
	Enabled *bool `json:"enabled,omitempty"`

	// Embedded configures the operator-managed Postgres instance.
	// Only used when Enabled is true.
	Embedded *EmbeddedPostgres `json:"embedded,omitempty"`

	// External references an existing Postgres instance.
	// Only used when Enabled is false.
	External *ExternalPostgres `json:"external,omitempty"`
}

// EmbeddedPostgres defines configuration for an operator-managed Postgres
// StatefulSet created per AxonHub instance.
type EmbeddedPostgres struct {
	// Image is the Postgres container image.
	// +kubebuilder:default="postgres:16"
	Image string `json:"image,omitempty"`

	// Database is the name of the database to create.
	// +kubebuilder:default="axonhub"
	Database string `json:"database,omitempty"`

	// User is the Postgres user that AxonHub will connect as.
	// +kubebuilder:default="axonhub"
	User string `json:"user,omitempty"`

	// PasswordSecretRef references a Secret containing the Postgres password.
	// If empty, the operator generates a random password and stores it in a
	// Secret named "<axonhub-name>-pg-password".
	PasswordSecretRef *corev1.SecretKeySelector `json:"passwordSecretRef,omitempty"`

	// Storage is the size of the persistent volume for Postgres data.
	// +kubebuilder:default="10Gi"
	Storage string `json:"storage,omitempty"`

	// ShmSize is the size of the shared memory (/dev/shm) for Postgres.
	// +kubebuilder:default="256Mi"
	ShmSize string `json:"shmSize,omitempty"`

	// MaxConnections is the PostgreSQL max_connections setting.
	// +kubebuilder:default=512
	MaxConnections int32 `json:"maxConnections,omitempty"`

	// SharedBuffers is the PostgreSQL shared_buffers setting.
	// +kubebuilder:default="128MB"
	SharedBuffers string `json:"sharedBuffers,omitempty"`

	// Resources defines compute resources for the Postgres container.
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// ExternalPostgres defines connection details for an existing external Postgres
// instance that the operator will not manage.
type ExternalPostgres struct {
	// Host is the Postgres host address.
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// Port is the Postgres port.
	// +kubebuilder:default=5432
	Port int32 `json:"port,omitempty"`

	// Database is the name of the database to connect to.
	// +kubebuilder:default="axonhub"
	Database string `json:"database,omitempty"`

	// User is the Postgres user.
	// +kubebuilder:default="axonhub"
	User string `json:"user,omitempty"`

	// PasswordSecretRef references a Secret containing the Postgres password.
	// +kubebuilder:validation:Required
	PasswordSecretRef *corev1.SecretKeySelector `json:"passwordSecretRef"`

	// SSLMode controls the SSL connection mode for the Postgres connection.
	// Common values: disable, require, verify-ca, verify-full.
	// +kubebuilder:default="disable"
	SSLMode string `json:"sslMode,omitempty"`
}

// AxonHubStatus defines the observed state of AxonHub.
type AxonHubStatus struct {
	// Conditions represent the latest available observations of the AxonHub state.
	// +kubebuilder:patchStrategy=merge
	// +kubebuilder:patchMergeKey=type
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Ready indicates whether the AxonHub instance is ready to serve traffic.
	Ready bool `json:"ready,omitempty"`

	// DatabaseReady indicates whether the database (embedded or external) is ready.
	DatabaseReady bool `json:"databaseReady,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AxonHub is the Schema for the axonhubs API.
// +kubebuilder:resource:shortName=ah
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas"
// +kubebuilder:printcolumn:name="DB Ready",type="string",JSONPath=".status.databaseReady"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type AxonHub struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AxonHubSpec   `json:"spec,omitempty"`
	Status AxonHubStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AxonHubList contains a list of AxonHub.
type AxonHubList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AxonHub `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AxonHub{}, &AxonHubList{})
}
