package controllers

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type NamespaceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Namespace string
}

const (
	secretsAnnotation = "namespace-provisioner.jeckscolari.github.com/secrets"
)

//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete

func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var ns corev1.Namespace
	if err := r.Get(ctx, req.NamespacedName, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch namespace")
		return ctrl.Result{}, err
	}

	if ns.Status.Phase == corev1.NamespaceActive {

		if secretNames, ok := ns.ObjectMeta.Annotations[secretsAnnotation]; ok {
			r.createSecrets(ctx, secretNames, &ns)
		}
	}

	return ctrl.Result{}, nil
}

func (r *NamespaceReconciler) createSecrets(ctx context.Context, secretNames string, ns *corev1.Namespace) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	for _, secretName := range strings.Split(secretNames, ",") {

		var secret corev1.Secret
		if err := r.Client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: r.Namespace}, &secret); err != nil {
			if apierrors.IsNotFound(err) {
				log.Error(err, "Unable to fetch secret")
			}
		}

		secret = corev1.Secret{
			Data:       secret.Data,
			StringData: secret.StringData,
			Type:       secret.Type,
		}

		secret.Name = secretName
		secret.Namespace = ns.Name

		log.Info("Creating secret in namespace")

		if err := r.Client.Create(ctx, &secret); err != nil {
			if apierrors.IsAlreadyExists(err) {
				log.Error(err, "Secret already exists")
				return ctrl.Result{}, nil
			}
			log.Error(err, "Unknown error")
			return ctrl.Result{}, err
		}

		log.Info("Secret created")
	}

	return ctrl.Result{}, nil

}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}
