package servicebindingrequest

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// Add creates a new ServiceBindingRequest Controller and adds it to the Manager. The Manager will
// set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	dynClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, err
	}
	return &Reconciler{client: mgr.GetClient(), dynClient: dynClient, scheme: mgr.GetScheme()}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler.
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	opts := controller.Options{Reconciler: r}
	c, err := controller.New("servicebindingrequest-controller", mgr, opts)
	if err != nil {
		return err
	}

	enqueue := &handler.EnqueueRequestsFromMapFunc{ToRequests: &SBRRequestMapper{}}
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// ignore updates to CR status in which case metadata.Generation does not change
			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// evaluates to false if the object has been confirmed deleted
			return !e.DeleteStateUnknown
		},
	}

	for _, gvk := range getWatchingGVKs() {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		// TODO: add logging to show which GVKs this controller is considering;
		if err = c.Watch(&source.Kind{Type: u}, enqueue, pred); err != nil {
			return err
		}
	}

	return nil
}

func getWatchingGVKs() []schema.GroupVersionKind {
	return []schema.GroupVersionKind{
		v1alpha1.SchemeGroupVersion.WithKind("ServiceBindingRequest"),
		{Group: "", Version: "v1", Kind: "Secret"},
	}
}

// blank assignment to verify that ReconcileServiceBindingRequest implements reconcile.Reconciler
var _ reconcile.Reconciler = &Reconciler{}
