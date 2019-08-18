package servicebindingrequest

import (
	"context"
	"fmt"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/planner"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new ServiceBindingRequest Controller and adds it to the Manager. The Manager will
// set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	r := newReconciler(mgr)
	return add(mgr, newReconciler(mgr), r.ReconcileIfAssociatedWithAServiceBinding)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) *Reconciler {
	return &Reconciler{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler, nonServiceBindingOwnedTrigger handler.ToRequestsFunc) error {
	// Create a new controller
	c, err := controller.New("servicebindingrequest-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}

	// Watch for changes to primary resource ServiceBindingRequest
	err = c.Watch(&source.Kind{Type: &v1alpha1.ServiceBindingRequest{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	handlerSecret := &handler.EnqueueRequestsFromMapFunc{
		ToRequests: nonServiceBindingOwnedTrigger,
	}
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, handlerSecret)
	if err != nil {
		return err
	}

	handlerConfigMap := &handler.EnqueueRequestsFromMapFunc{
		ToRequests: nonServiceBindingOwnedTrigger,
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, handlerConfigMap)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) ReconcileIfAssociatedWithAServiceBinding(o handler.MapObject) []reconcile.Request {
	var result []reconcile.Request

	var objOwner *metav1.OwnerReference
	for _, owner := range o.Meta.GetOwnerReferences() {
		objOwner = &owner
		if owner.Kind == "ServiceBindingRequest" {
			fmt.Println("ConfigMap/Secret is managed by ServiceBindingRequest, dropping event")
			return nil
		}
	}

	sbr := &v1alpha1.ServiceBindingRequestList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingRequest",
			APIVersion: "apps.openshift.io/v1alpha1",
		},
	}
	// Get all ServiceBindingRequests
	if err := r.client.List(context.TODO(), nil, sbr); err != nil {
		return result
	}

	for _, sbr := range sbr.Items {
		// if the secret/configmap is owned to a CR which was bound in
		// a ServiceBindingRequest previously, reconcile is needed.

		if objOwner != nil {
			if sbr.Spec.BackingServiceSelector.ResourceRef == objOwner.Name &&
				sbr.Spec.BackingServiceSelector.Kind == objOwner.Kind {

				result = append(result, reconcile.Request{
					NamespacedName: client.ObjectKey{Namespace: sbr.Namespace, Name: sbr.Name}})
			}

		} else {
			plannerRef := planner.NewPlanner(context.TODO(), r.client, o.Meta.GetNamespace(), &sbr)
			plan, err := plannerRef.Plan()
			if err != nil {
				continue
			}
			retrieverObj := NewRetriever(context.TODO(), r.client, plan, "")
			if o.Object.GetObjectKind().GroupVersionKind().Kind == "Secret" {
				for _, s := range retrieverObj.secrets {
					// if it happens to be one of the secrets consumed
					// by the CR in the spec but not necessarily owned
					// by the CR.
					if s.Name == o.Meta.GetName() {
						result = append(result, reconcile.Request{
							NamespacedName: client.ObjectKey{Namespace: sbr.Namespace, Name: sbr.Name}})
						break
					}
				}
			} else if o.Object.GetObjectKind().GroupVersionKind().Kind == "ConfigMap" {
				for _, s := range retrieverObj.configmaps {
					// if it happens to be one of the configmaps
					// consumed by the CR in the spec
					// but not necessarily owned by the CR.
					if s.Name == o.Meta.GetName() {
						result = append(result, reconcile.Request{
							NamespacedName: client.ObjectKey{Namespace: sbr.Namespace, Name: sbr.Name}})
						break
					}
				}
			}
		}
	}
	return result
}

/*
// NonServiceBindingOwnedCOnfigMapTrigger is a trigger on all secrets in that namespace
func (r *Reconciler) NonServiceBindingOwnedCOnfigMapTrigger(o handler.MapObject) []reconcile.Request {
	for _, owner := range o.Meta.GetOwnerReferences() {
		if owner.Name == "" {
			// if the owner is not present, we are not really concerned.
			return nil
		}
		if owner.Kind == "ServiceBindingRequest" {
			fmt.Println("ConfigMap/Secret is managed by ServiceBindingRequest, dropping event")
			return nil
		}
	}

	// Fetch the triggered ConfigMap Data
	configMapInstance := &corev1.ConfigMap{}
	key := client.ObjectKey{Namespace: o.Meta.GetNamespace(), Name: o.Meta.GetName()}
	err := r.client.Get(context.TODO(), key, configMapInstance)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return nil
	}

	return r.reconcileIfAssociatedWithAServiceBinding(o)
}

// NonServiceBindingOwnedSecretTrigger is a trigger on all secrets in that namespace
func (r *Reconciler) NonServiceBindingOwnedSecretTrigger(o handler.MapObject) []reconcile.Request {

	for _, owner := range o.Meta.GetOwnerReferences() {
		if owner.Name == "" {
			// if the owner is not present, we are not really concerned.
			return nil
		}
		if owner.Kind == "ServiceBindingRequest" {
			fmt.Println("Secret is managed by ServiceBindingRequest, dropping event")
			return nil
		}
	}

	// Fetch the triggered Secret Data
	instance := &corev1.Secret{}
	key := client.ObjectKey{Namespace: o.Meta.GetNamespace(), Name: o.Meta.GetName()}
	err := r.client.Get(context.TODO(), key, instance)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return nil
	}

	return r.reconcileIfAssociatedWithAServiceBinding(o)
}
*/

// blank assignment to verify that ReconcileServiceBindingRequest implements reconcile.Reconciler
var _ reconcile.Reconciler = &Reconciler{}
