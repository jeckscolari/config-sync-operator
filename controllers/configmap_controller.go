/*
Copyright 2021.

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

package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ConfigMapReconciler reconciles a ConfigMap object
type ConfigMapReconciler struct {
	client.Client
}

//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=configmaps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ConfigMap object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *ConfigMapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var configMap corev1.ConfigMap

	if err := r.Get(ctx, req.NamespacedName, &configMap); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch ConfigMap")
		return ctrl.Result{}, err
	}

	log.Info("processing ConfigMap")

	namespacesSelector := configMap.Annotations[SyncAnnotation]

	if len(namespacesSelector) > 0 {
		namespaces, err := ListNamespaces(r.Client, ctx, namespacesSelector)

		if err != nil {
			log.Error(err, "unable to list Namespaces")
		}

		for _, namespace := range namespaces.Items {
			r.upsertConfigMap(ctx, configMap, namespace)
		}
	}

	return ctrl.Result{}, nil
}

func (r *ConfigMapReconciler) upsertConfigMap(ctx context.Context, source corev1.ConfigMap, target corev1.Namespace) error {
	log := log.FromContext(ctx)

	meta := metav1.ObjectMeta{
		Name:      source.Name,
		Namespace: target.Name,
	}

	var old corev1.ConfigMap

	if err := r.Get(ctx, types.NamespacedName{Name: source.Name, Namespace: target.Name}, &old); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("creating ConfigMap", "source object", source.Name, "target namespace", target.Name)

			if err := r.Create(ctx, &corev1.ConfigMap{
				ObjectMeta: meta,
				Data:       source.Data,
			}); err != nil {
				log.Error(err, "unable to create ConfigMap")
			}
		} else {
			return err
		}
	} else {
		log.Info("updating ConfigMap", "source object", source.Name, "target namespace", target.Name)

		old.Data = source.Data

		if err := r.Update(ctx, &old); err != nil {
			log.Error(err, "unable to update ConfigMap")
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigMapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(r)
}
