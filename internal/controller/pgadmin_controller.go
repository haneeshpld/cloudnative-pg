package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	postgresqlv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
)

// PgAdminReconciler reconciles a PgAdmin object
type PgAdminReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=postgresql.cnpg.io,resources=pgadmins,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=postgresql.cnpg.io,resources=pgadmins/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=postgresql.cnpg.io,resources=pgadmins/finalizers,verbs=update

func (r *PgAdminReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("🚨 PgAdmin Reconciler called", "name", req.Name, "namespace", req.Namespace)

	var pgAdmin postgresqlv1.PgAdmin
	if err := r.Get(ctx, req.NamespacedName, &pgAdmin); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	dep := newDeployment(&pgAdmin)
	if err := ctrl.SetControllerReference(&pgAdmin, dep, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	var foundDep appsv1.Deployment
	if err := r.Get(ctx, client.ObjectKey{Name: dep.Name, Namespace: dep.Namespace}, &foundDep); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("Creating Deployment", "namespace", dep.Namespace, "name", dep.Name)
			if err := r.Create(ctx, dep); err != nil {
				logger.Error(err, "Failed to create Deployment", "namespace", dep.Namespace, "name", dep.Name)
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	}

	svc := newService(&pgAdmin)
	if err := ctrl.SetControllerReference(&pgAdmin, svc, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	var foundSvc corev1.Service
	if err := r.Get(ctx, client.ObjectKey{Name: svc.Name, Namespace: svc.Namespace}, &foundSvc); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("Creating Service", "namespace", svc.Namespace, "name", svc.Name)
			if err := r.Create(ctx, svc); err != nil {
				logger.Error(err, "Failed to create Service", "namespace", svc.Namespace, "name", svc.Name)
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *PgAdminReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&postgresqlv1.PgAdmin{}).
		Complete(r)
}

// newDeployment returns a Deployment object for pgAdmin.
func newDeployment(cr *postgresqlv1.PgAdmin) *appsv1.Deployment {
	labels := map[string]string{"app": "pgadmin"}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pgadmin-deployment",
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: func(b bool) *bool { return &b }(false),
					},
					Containers: []corev1.Container{
						{
							Name:  "pgadmin",
							Image: "ghcr.io/haneeshpld/pgadmin4-nonroot:latest",
							Env: []corev1.EnvVar{
								{
									Name:  "PGADMIN_DEFAULT_EMAIL",
									Value: "admin@example.com",
								},
								{
									Name:  "PGADMIN_DEFAULT_PASSWORD",
									Value: "admin",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								RunAsUser:                int64Ptr(0),
								AllowPrivilegeEscalation: boolPtr(true),
							},
						},
					},
				},
			},
		},
	}
}

// newService returns a Service object for pgAdmin.
func newService(cr *postgresqlv1.PgAdmin) *corev1.Service {
	labels := map[string]string{"app": "pgadmin"}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pgadmin-service",
			Namespace: cr.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
		},
	}
}

// Helper functions
func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool    { return &b }
