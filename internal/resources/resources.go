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

package resources

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	axonhubv1alpha1 "github.com/amghazanfari/axon-operator/api/v1alpha1"
)

const (
	// LabelKeyManagedBy is the label key used to identify resources managed by this operator.
	LabelKeyManagedBy = "app.kubernetes.io/managed-by"
	// LabelKeyInstance is the label key for the AxonHub instance name.
	LabelKeyInstance = "app.kubernetes.io/instance"
	// LabelKeyComponent is the label key for the component (axonhub or postgres).
	LabelKeyComponent = "app.kubernetes.io/component"
	// LabelValueManagedBy is the value for the managed-by label.
	LabelValueManagedBy = "axon-operator"
	// ComponentAxonHub is the component label value for AxonHub.
	ComponentAxonHub = "axonhub"
	// ComponentPostgres is the component label value for Postgres.
	ComponentPostgres = "postgres"

	// DefaultPgPort is the default Postgres port.
	DefaultPgPort = 5432
	// DefaultAxonHubPort is the default AxonHub port.
	DefaultAxonHubPort = 8090
	// DefaultPgImage is the default Postgres image.
	DefaultPgImage = "postgres:16"
	// DefaultPgDatabase is the default Postgres database name.
	DefaultPgDatabase = "axonhub"
	// DefaultPgUser is the default Postgres user.
	DefaultPgUser = "axonhub"
	// DefaultPgStorage is the default Postgres storage size.
	DefaultPgStorage = "10Gi"
	// DefaultPgShmSize is the default Postgres shared memory size.
	DefaultPgShmSize = "256Mi"
	// DefaultPgMaxConnections is the default Postgres max_connections.
	DefaultPgMaxConnections = 512
	// DefaultPgSharedBuffers is the default Postgres shared_buffers.
	DefaultPgSharedBuffers = "128MB"
)

// CommonLabels returns the labels applied to all resources created for an AxonHub instance.
func CommonLabels(axonhub *axonhubv1alpha1.AxonHub) map[string]string {
	return map[string]string{
		LabelKeyManagedBy: LabelValueManagedBy,
		LabelKeyInstance:  axonhub.Name,
	}
}

// AxonHubLabels returns labels for AxonHub application resources.
func AxonHubLabels(axonhub *axonhubv1alpha1.AxonHub) map[string]string {
	labels := CommonLabels(axonhub)
	labels[LabelKeyComponent] = ComponentAxonHub
	return labels
}

// PostgresLabels returns labels for Postgres resources.
func PostgresLabels(axonhub *axonhubv1alpha1.AxonHub) map[string]string {
	labels := CommonLabels(axonhub)
	labels[LabelKeyComponent] = ComponentPostgres
	return labels
}

// PgSecretName returns the name of the Postgres password secret for the given AxonHub.
func PgSecretName(axonhub *axonhubv1alpha1.AxonHub) string {
	return axonhub.Name + "-pg-password"
}

// PgStatefulSetName returns the name of the Postgres StatefulSet for the given AxonHub.
func PgStatefulSetName(axonhub *axonhubv1alpha1.AxonHub) string {
	return axonhub.Name + "-postgres"
}

// PgServiceName returns the name of the Postgres Service for the given AxonHub.
func PgServiceName(axonhub *axonhubv1alpha1.AxonHub) string {
	return axonhub.Name + "-postgres"
}

// AxonHubServiceName returns the name of the AxonHub Service for the given AxonHub.
func AxonHubServiceName(axonhub *axonhubv1alpha1.AxonHub) string {
	return axonhub.Name
}

// AxonHubDeploymentName returns the name of the AxonHub Deployment for the given AxonHub.
func AxonHubDeploymentName(axonhub *axonhubv1alpha1.AxonHub) string {
	return axonhub.Name
}

// PostgresConfig holds resolved Postgres configuration with defaults applied.
type PostgresConfig struct {
	Image          string
	Database       string
	User           string
	Storage        string
	ShmSize        string
	MaxConnections int32
	SharedBuffers  string
	Resources      *corev1.ResourceRequirements
}

// ResolveEmbeddedPostgresConfig resolves the embedded Postgres config with defaults.
func ResolveEmbeddedPostgresConfig(embedded *axonhubv1alpha1.EmbeddedPostgres) PostgresConfig {
	cfg := PostgresConfig{
		Image:          DefaultPgImage,
		Database:       DefaultPgDatabase,
		User:           DefaultPgUser,
		Storage:        DefaultPgStorage,
		ShmSize:        DefaultPgShmSize,
		MaxConnections: DefaultPgMaxConnections,
		SharedBuffers:  DefaultPgSharedBuffers,
	}
	if embedded != nil {
		if embedded.Image != "" {
			cfg.Image = embedded.Image
		}
		if embedded.Database != "" {
			cfg.Database = embedded.Database
		}
		if embedded.User != "" {
			cfg.User = embedded.User
		}
		if embedded.Storage != "" {
			cfg.Storage = embedded.Storage
		}
		if embedded.ShmSize != "" {
			cfg.ShmSize = embedded.ShmSize
		}
		if embedded.MaxConnections != 0 {
			cfg.MaxConnections = embedded.MaxConnections
		}
		if embedded.SharedBuffers != "" {
			cfg.SharedBuffers = embedded.SharedBuffers
		}
		cfg.Resources = embedded.Resources
	}
	return cfg
}

// PgConnectionString returns the DATABASE_URL-style connection string env vars for AxonHub.
func PgConnectionString(axonhub *axonhubv1alpha1.AxonHub, host string, port int32, database, user string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "POSTGRES_HOST",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
	}
}

// BuildPgSecret builds the Postgres password Secret. If passwordSecretRef is provided,
// the operator does not create a secret (the user-provided one is used).
func BuildPgSecret(axonhub *axonhubv1alpha1.AxonHub, password string, scheme *runtime.Scheme) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PgSecretName(axonhub),
			Namespace: axonhub.Namespace,
			Labels:    PostgresLabels(axonhub),
		},
		StringData: map[string]string{
			"password": password,
		},
	}
	return secret
}

// BuildPgService builds the Postgres headless Service for the StatefulSet.
func BuildPgService(axonhub *axonhubv1alpha1.AxonHub) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PgServiceName(axonhub),
			Namespace: axonhub.Namespace,
			Labels:    PostgresLabels(axonhub),
		},
		Spec: corev1.ServiceSpec{
			Selector:  PostgresLabels(axonhub),
			ClusterIP: corev1.ClusterIPNone,
			Ports: []corev1.ServicePort{
				{
					Name:       "postgres",
					Port:       DefaultPgPort,
					TargetPort: intstr.FromInt(DefaultPgPort),
				},
			},
		},
	}
}

// BuildPgStatefulSet builds the Postgres StatefulSet for embedded Postgres.
func BuildPgStatefulSet(axonhub *axonhubv1alpha1.AxonHub, scheme *runtime.Scheme) *appsv1.StatefulSet {
	cfg := ResolveEmbeddedPostgresConfig(axonhub.Spec.Postgres.Embedded)

	shmSize := resource.MustParse(cfg.ShmSize)

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PgStatefulSetName(axonhub),
			Namespace: axonhub.Namespace,
			Labels:    PostgresLabels(axonhub),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: PgServiceName(axonhub),
			Replicas:    ptr(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: PostgresLabels(axonhub),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: PostgresLabels(axonhub),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "postgres",
							Image: cfg.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: DefaultPgPort,
									Name:          "postgres",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "POSTGRES_DB",
									Value: cfg.Database,
								},
								{
									Name:  "POSTGRES_USER",
									Value: cfg.User,
								},
								{
									Name: "POSTGRES_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: pgPasswordSecretRef(axonhub),
									},
								},
								{
									Name:  "PGDATA",
									Value: "/var/lib/postgresql/data/pgdata",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "postgres-data",
									MountPath: "/var/lib/postgresql/data",
								},
								{
									Name:      "dshm",
									MountPath: "/dev/shm",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: shmSize,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dshm",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    corev1.StorageMediumMemory,
									SizeLimit: &shmSize,
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "postgres-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(cfg.Storage),
							},
						},
					},
				},
			},
		},
	}

	if cfg.Resources != nil {
		sts.Spec.Template.Spec.Containers[0].Resources = *cfg.Resources
	}

	return sts
}

// pgPasswordSecretRef returns the SecretKeySelector for the Postgres password.
// If the user provided a PasswordSecretRef, use it; otherwise use the operator-managed secret.
func pgPasswordSecretRef(axonhub *axonhubv1alpha1.AxonHub) *corev1.SecretKeySelector {
	if axonhub.Spec.Postgres.Embedded != nil && axonhub.Spec.Postgres.Embedded.PasswordSecretRef != nil {
		return axonhub.Spec.Postgres.Embedded.PasswordSecretRef
	}
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: PgSecretName(axonhub),
		},
		Key: "password",
	}
}

// BuildAxonHubService builds the Service for the AxonHub application.
func BuildAxonHubService(axonhub *axonhubv1alpha1.AxonHub) *corev1.Service {
	port := axonhub.Spec.Port
	if port == 0 {
		port = DefaultAxonHubPort
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AxonHubServiceName(axonhub),
			Namespace: axonhub.Namespace,
			Labels:    AxonHubLabels(axonhub),
		},
		Spec: corev1.ServiceSpec{
			Selector: AxonHubLabels(axonhub),
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       port,
					TargetPort: intstr.FromInt(int(port)),
				},
			},
		},
	}
}

// BuildAxonHubDeployment builds the Deployment for the AxonHub application.
func BuildAxonHubDeployment(axonhub *axonhubv1alpha1.AxonHub) *appsv1.Deployment {
	replicas := int32(1)
	if axonhub.Spec.Replicas != nil {
		replicas = *axonhub.Spec.Replicas
	}

	port := axonhub.Spec.Port
	if port == 0 {
		port = DefaultAxonHubPort
	}

	envVars := []corev1.EnvVar{}
	if axonhub.Spec.Postgres.Enabled == nil || *axonhub.Spec.Postgres.Enabled {
		envVars = append(envVars, buildEmbeddedPgEnv(axonhub)...)
	} else if axonhub.Spec.Postgres.External != nil {
		envVars = append(envVars, buildExternalPgEnv(axonhub.Spec.Postgres.External)...)
	}
	envVars = append(envVars, axonhub.Spec.Env...)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AxonHubDeploymentName(axonhub),
			Namespace: axonhub.Namespace,
			Labels:    AxonHubLabels(axonhub),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: AxonHubLabels(axonhub),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: AxonHubLabels(axonhub),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "axonhub",
							Image: axonhub.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: port,
									Name:          "http",
								},
							},
							Env: envVars,
						},
					},
				},
			},
		},
	}

	if axonhub.Spec.Resources != nil {
		dep.Spec.Template.Spec.Containers[0].Resources = *axonhub.Spec.Resources
	}

	return dep
}

func buildEmbeddedPgEnv(axonhub *axonhubv1alpha1.AxonHub) []corev1.EnvVar {
	cfg := ResolveEmbeddedPostgresConfig(axonhub.Spec.Postgres.Embedded)
	pgHost := fmt.Sprintf("%s.%s.svc.cluster.local", PgServiceName(axonhub), axonhub.Namespace)
	return []corev1.EnvVar{
		{Name: "POSTGRES_HOST", Value: pgHost},
		{Name: "POSTGRES_PORT", Value: fmt.Sprintf("%d", DefaultPgPort)},
		{Name: "POSTGRES_DB", Value: cfg.Database},
		{Name: "POSTGRES_USER", Value: cfg.User},
		{
			Name: "POSTGRES_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: pgPasswordSecretRef(axonhub),
			},
		},
	}
}

func buildExternalPgEnv(ext *axonhubv1alpha1.ExternalPostgres) []corev1.EnvVar {
	port := ext.Port
	if port == 0 {
		port = DefaultPgPort
	}
	database := ext.Database
	if database == "" {
		database = DefaultPgDatabase
	}
	user := ext.User
	if user == "" {
		user = DefaultPgUser
	}
	return []corev1.EnvVar{
		{Name: "POSTGRES_HOST", Value: ext.Host},
		{Name: "POSTGRES_PORT", Value: fmt.Sprintf("%d", port)},
		{Name: "POSTGRES_DB", Value: database},
		{Name: "POSTGRES_USER", Value: user},
		{
			Name: "POSTGRES_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: ext.PasswordSecretRef,
			},
		},
	}
}

func ptr[T any](v T) *T {
	return &v
}
