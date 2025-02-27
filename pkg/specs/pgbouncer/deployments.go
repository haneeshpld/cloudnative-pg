/*
Copyright The CloudNativePG Contributors

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

// Package pgbouncer contains the specification of the K8s resources
// generated by the CloudNativePG operator related to pgbouncer poolers
package pgbouncer

import (
	"path"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	apiv1 "github.com/haneeshpld/cloudnative-pg/api/v1"
	config "github.com/haneeshpld/cloudnative-pg/internal/configuration"
	pgBouncerConfig "github.com/haneeshpld/cloudnative-pg/pkg/management/pgbouncer/config"
	"github.com/haneeshpld/cloudnative-pg/pkg/management/url"
	"github.com/haneeshpld/cloudnative-pg/pkg/podspec"
	"github.com/haneeshpld/cloudnative-pg/pkg/postgres"
	"github.com/haneeshpld/cloudnative-pg/pkg/specs"
	"github.com/haneeshpld/cloudnative-pg/pkg/utils"
	"github.com/haneeshpld/cloudnative-pg/pkg/utils/hash"
)

const (
	// DefaultPgbouncerImage is the name of the pgbouncer image used by default
	DefaultPgbouncerImage = "ghcr.io/cloudnative-pg/pgbouncer:1.24.0"
)

// Deployment create the deployment of pgbouncer, given
// the configurations we have in the pooler specifications
func Deployment(pooler *apiv1.Pooler, cluster *apiv1.Cluster) (*appsv1.Deployment, error) {
	operatorImageName := config.Current.OperatorImageName

	poolerHash, err := computeTemplateHash(pooler, operatorImageName)
	if err != nil {
		return nil, err
	}

	podTemplate := podspec.NewFrom(pooler.Spec.Template).
		WithLabel(utils.PgbouncerNameLabel, pooler.Name).
		WithLabel(utils.ClusterLabelName, cluster.Name).
		WithLabel(utils.PodRoleLabelName, string(utils.PodRolePooler)).
		WithVolume(&corev1.Volume{
			Name: "ca",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: cluster.GetServerCASecretName(),
				},
			},
		}).
		WithVolume(&corev1.Volume{
			Name: "server-tls",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: cluster.GetServerTLSSecretName(),
				},
			},
		}).
		WithSecurityContext(specs.CreatePodSecurityContext(cluster.GetSeccompProfile(), 998, 996), true).
		WithContainerImage("pgbouncer", DefaultPgbouncerImage, false).
		WithContainerCommand("pgbouncer", []string{
			"/controller/manager",
			"pgbouncer",
			"run",
		}, false).
		WithContainerPort("pgbouncer", &corev1.ContainerPort{
			Name:          pgBouncerConfig.PgBouncerPortName,
			ContainerPort: pgBouncerConfig.PgBouncerPort,
		}).
		WithContainerPort("pgbouncer", &corev1.ContainerPort{
			Name:          "metrics",
			ContainerPort: url.PgBouncerMetricsPort,
		}).
		WithInitContainerImage(specs.BootstrapControllerContainerName, operatorImageName, true).
		WithInitContainerCommand(specs.BootstrapControllerContainerName,
			[]string{"/manager", "bootstrap", "/controller/manager"},
			true).
		WithInitContainerSecurityContext(specs.BootstrapControllerContainerName,
			specs.CreateContainerSecurityContext(cluster.GetSeccompProfile()),
			true).
		WithVolume(&corev1.Volume{
			Name: "scratch-data",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}).
		WithInitContainerVolumeMount(specs.BootstrapControllerContainerName, &corev1.VolumeMount{
			Name:      "scratch-data",
			MountPath: postgres.ScratchDataDirectory,
		}, true).
		WithContainerVolumeMount("pgbouncer", &corev1.VolumeMount{
			Name:      "scratch-data",
			MountPath: postgres.ScratchDataDirectory,
		}, true).
		WithContainerEnv("pgbouncer", corev1.EnvVar{Name: "NAMESPACE", Value: pooler.Namespace}, true).
		WithContainerEnv("pgbouncer", corev1.EnvVar{Name: "POOLER_NAME", Value: pooler.Name}, true).
		WithContainerEnv("pgbouncer", corev1.EnvVar{Name: "PGUSER", Value: "pgbouncer"}, false).
		WithContainerEnv("pgbouncer", corev1.EnvVar{Name: "PGDATABASE", Value: "pgbouncer"}, false).
		WithContainerEnv("pgbouncer", corev1.EnvVar{Name: "PGHOST", Value: "/controller/run"}, false).
		WithContainerEnv("pgbouncer", corev1.EnvVar{
			Name:  "PSQL_HISTORY",
			Value: path.Join(postgres.TemporaryDirectory, ".psql_history"),
		}, false).
		WithContainerSecurityContext("pgbouncer", specs.CreateContainerSecurityContext(cluster.GetSeccompProfile()), true).
		WithServiceAccountName(pooler.Name, true).
		WithReadinessProbe("pgbouncer", &corev1.Probe{
			TimeoutSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt32(pgBouncerConfig.PgBouncerPort),
				},
			},
		}, false).
		Build()

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pooler.Name,
			Namespace: pooler.Namespace,
			Labels: map[string]string{
				utils.ClusterLabelName:   cluster.Name,
				utils.PgbouncerNameLabel: pooler.Name,
				utils.PodRoleLabelName:   string(utils.PodRolePooler),
			},
			Annotations: map[string]string{
				utils.PoolerSpecHashAnnotationName: poolerHash,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pooler.Spec.Instances,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					utils.PgbouncerNameLabel: pooler.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: podTemplate.ObjectMeta.Annotations,
					Labels:      podTemplate.ObjectMeta.Labels,
				},
				Spec: podTemplate.Spec,
			},
			Strategy: getDeploymentStrategy(pooler.Spec.DeploymentStrategy),
		},
	}, nil
}

func computeTemplateHash(pooler *apiv1.Pooler, operatorImageName string) (string, error) {
	type deploymentHash struct {
		poolerSpec                      apiv1.PoolerSpec
		operatorImageName               string
		isPodSpecReconciliationDisabled bool
	}

	return hash.ComputeHash(deploymentHash{
		poolerSpec:                      pooler.Spec,
		operatorImageName:               operatorImageName,
		isPodSpecReconciliationDisabled: utils.IsPodSpecReconciliationDisabled(&pooler.ObjectMeta),
	})
}

func getDeploymentStrategy(strategy *appsv1.DeploymentStrategy) appsv1.DeploymentStrategy {
	if strategy != nil {
		return *strategy.DeepCopy()
	}
	return appsv1.DeploymentStrategy{}
}
