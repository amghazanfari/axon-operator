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

package controller

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	axonhubv1alpha1 "github.com/amghazanfari/axon-operator/api/v1alpha1"
	"github.com/amghazanfari/axon-operator/internal/resources"
)

const (
	finalizerName = "axonhub.looplj.com/finalizer"

	conditionTypeReady      = "Ready"
	conditionTypeDatabase   = "DatabaseReady"
	conditionTypeDeployment = "DeploymentReady"

	reasonReconciling   = "Reconciling"
	reasonReconciled    = "Reconciled"
	reasonError         = "Error"
	reasonPostgresReady = "PostgresReady"
	reasonPostgresIssue = "PostgresIssue"
)

// AxonHubReconciler reconciles a AxonHub object
type AxonHubReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=axonhub.looplj.com,resources=axonhubs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=axonhub.looplj.com,resources=axonhubs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=axonhub.looplj.com,resources=axonhubs/finalizers,verbs=update

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AxonHubReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var axonhub axonhubv1alpha1.AxonHub
	if err := r.Get(ctx, req.NamespacedName, &axonhub); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("AxonHub resource not found, ignoring")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get AxonHub resource")
		return ctrl.Result{}, err
	}

	if !axonhub.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, &axonhub)
	}

	return r.reconcileNormal(ctx, &axonhub)
}

func (r *AxonHubReconciler) reconcileNormal(ctx context.Context, axonhub *axonhubv1alpha1.AxonHub) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(axonhub, finalizerName) {
		controllerutil.AddFinalizer(axonhub, finalizerName)
		if err := r.Update(ctx, axonhub); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	postgresReady := true

	if axonhub.Spec.Postgres.Enabled == nil || *axonhub.Spec.Postgres.Enabled {
		pgReady, err := r.reconcileEmbeddedPostgres(ctx, axonhub)
		if err != nil {
			r.setCondition(ctx, axonhub, conditionTypeDatabase, metav1.ConditionFalse, reasonError, err.Error())
			return ctrl.Result{}, err
		}
		postgresReady = pgReady
	}

	if postgresReady {
		r.setCondition(ctx, axonhub, conditionTypeDatabase, metav1.ConditionTrue, reasonPostgresReady, "Postgres is ready")
	} else {
		r.setCondition(ctx, axonhub, conditionTypeDatabase, metav1.ConditionFalse, reasonPostgresIssue, "Waiting for Postgres to become ready")
	}

	depReady, err := r.reconcileAxonHubDeployment(ctx, axonhub)
	if err != nil {
		r.setCondition(ctx, axonhub, conditionTypeDeployment, metav1.ConditionFalse, reasonError, err.Error())
		return ctrl.Result{}, err
	}

	if depReady {
		r.setCondition(ctx, axonhub, conditionTypeDeployment, metav1.ConditionTrue, reasonReconciled, "Deployment is ready")
	} else {
		r.setCondition(ctx, axonhub, conditionTypeDeployment, metav1.ConditionFalse, reasonReconciling, "Waiting for Deployment to become ready")
	}

	allReady := postgresReady && depReady
	axonhub.Status.Ready = allReady
	axonhub.Status.DatabaseReady = postgresReady

	if allReady {
		r.setCondition(ctx, axonhub, conditionTypeReady, metav1.ConditionTrue, reasonReconciled, "AxonHub is ready")
	} else {
		r.setCondition(ctx, axonhub, conditionTypeReady, metav1.ConditionFalse, reasonReconciling, "AxonHub is not yet ready")
	}

	if err := r.Status().Update(ctx, axonhub); err != nil {
		logger.Error(err, "Failed to update AxonHub status")
		return ctrl.Result{}, err
	}

	if !allReady {
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *AxonHubReconciler) reconcileEmbeddedPostgres(ctx context.Context, axonhub *axonhubv1alpha1.AxonHub) (bool, error) {
	// Create or update the Postgres password secret if no user-provided ref
	if axonhub.Spec.Postgres.Embedded == nil || axonhub.Spec.Postgres.Embedded.PasswordSecretRef == nil {
		if err := r.reconcilePgSecret(ctx, axonhub); err != nil {
			return false, err
		}
	}

	// Create or update the Postgres Service
	pgSvc := resources.BuildPgService(axonhub)
	if err := controllerutil.SetControllerReference(axonhub, pgSvc, r.Scheme); err != nil {
		return false, err
	}
	if err := r.createOrUpdate(ctx, pgSvc); err != nil {
		return false, err
	}

	// Create or update the Postgres StatefulSet
	pgSts := resources.BuildPgStatefulSet(axonhub, r.Scheme)
	if err := controllerutil.SetControllerReference(axonhub, pgSts, r.Scheme); err != nil {
		return false, err
	}
	if err := r.createOrUpdate(ctx, pgSts); err != nil {
		return false, err
	}

	// Check readiness
	var sts appsv1.StatefulSet
	if err := r.Get(ctx, types.NamespacedName{Name: resources.PgStatefulSetName(axonhub), Namespace: axonhub.Namespace}, &sts); err != nil {
		return false, err
	}

	return sts.Status.ReadyReplicas >= 1, nil
}

func (r *AxonHubReconciler) reconcilePgSecret(ctx context.Context, axonhub *axonhubv1alpha1.AxonHub) error {
	secretName := resources.PgSecretName(axonhub)
	var existing corev1.Secret
	err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: axonhub.Namespace}, &existing)
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}

	password, err := generatePassword(32)
	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	secret := resources.BuildPgSecret(axonhub, password, r.Scheme)
	if err := controllerutil.SetControllerReference(axonhub, secret, r.Scheme); err != nil {
		return err
	}
	return r.Create(ctx, secret)
}

func (r *AxonHubReconciler) reconcileAxonHubDeployment(ctx context.Context, axonhub *axonhubv1alpha1.AxonHub) (bool, error) {
	// Create or update the AxonHub Service
	axSvc := resources.BuildAxonHubService(axonhub)
	if err := controllerutil.SetControllerReference(axonhub, axSvc, r.Scheme); err != nil {
		return false, err
	}
	if err := r.createOrUpdate(ctx, axSvc); err != nil {
		return false, err
	}

	// Create or update the AxonHub Deployment
	dep := resources.BuildAxonHubDeployment(axonhub)
	if err := controllerutil.SetControllerReference(axonhub, dep, r.Scheme); err != nil {
		return false, err
	}
	if err := r.createOrUpdate(ctx, dep); err != nil {
		return false, err
	}

	// Check readiness
	var deployment appsv1.Deployment
	if err := r.Get(ctx, types.NamespacedName{Name: resources.AxonHubDeploymentName(axonhub), Namespace: axonhub.Namespace}, &deployment); err != nil {
		return false, err
	}

	if deployment.Status.ReadyReplicas >= 1 {
		return true, nil
	}
	return false, nil
}

func (r *AxonHubReconciler) reconcileDelete(ctx context.Context, axonhub *axonhubv1alpha1.AxonHub) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Deleting AxonHub, cleaning up")

	controllerutil.RemoveFinalizer(axonhub, finalizerName)
	if err := r.Update(ctx, axonhub); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *AxonHubReconciler) createOrUpdate(ctx context.Context, obj client.Object) error {
	key := types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}

	current := obj.DeepCopyObject().(client.Object)
	err := r.Get(ctx, key, current)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		return r.Create(ctx, obj)
	}

	changed := false

	if !reflect.DeepEqual(obj.GetLabels(), current.GetLabels()) {
		current.SetLabels(obj.GetLabels())
		changed = true
	}

	switch desired := obj.(type) {
	case *corev1.Service:
		currentSvc := current.(*corev1.Service)
		desired.Spec.ClusterIP = currentSvc.Spec.ClusterIP
		if !reflect.DeepEqual(desired.Spec, currentSvc.Spec) {
			currentSvc.Spec = desired.Spec
			changed = true
		}
	case *appsv1.Deployment:
		currentDep := current.(*appsv1.Deployment)
		if !reflect.DeepEqual(desired.Spec, currentDep.Spec) {
			currentDep.Spec = desired.Spec
			changed = true
		}
	case *appsv1.StatefulSet:
		currentSts := current.(*appsv1.StatefulSet)
		if !reflect.DeepEqual(desired.Spec, currentSts.Spec) {
			currentSts.Spec = desired.Spec
			changed = true
		}
	case *corev1.Secret:
		currentSec := current.(*corev1.Secret)
		if !reflect.DeepEqual(desired.StringData, currentSec.StringData) {
			currentSec.StringData = desired.StringData
			changed = true
		}
	}

	if !changed {
		return nil
	}

	if err := r.Update(ctx, current); err != nil {
		if errors.IsConflict(err) {
			return nil
		}
		return err
	}

	return nil
}

func (r *AxonHubReconciler) setCondition(_ context.Context, axonhub *axonhubv1alpha1.AxonHub, condType string, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: axonhub.Generation,
	}
	meta.SetStatusCondition(&axonhub.Status.Conditions, condition)
}

func generatePassword(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AxonHubReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&axonhubv1alpha1.AxonHub{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Named("axonhub").
		Complete(r)
}
