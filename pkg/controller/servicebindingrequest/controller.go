package servicebindingrequest

import (
	"github.com/redhat-developer/service-binding-operator/pkg/controller/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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

func addServiceBindingRequestWatches(c controller.Controller) error {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(e.ObjectNew)
			if err != nil {
				// TODO: add logging to show this error;
				return false
			}

			u := &unstructured.Unstructured{Object: data}
			sbr := &v1alpha1.ServiceBindingRequest{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sbr)
			if err != nil {
				// TODO: add logging to show this error;
				return false
			}

			// allowing pending SBR to be reconciled anyways
			// FIXME: use the constant defined in catchall reconciler;
			if sbr.Status.BindingStatus == "pending" {
				return true
			}

			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}

	// Watch for changes to primary resource ServiceBindingRequest
	err := c.Watch(&source.Kind{Type: &v1alpha1.ServiceBindingRequest{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	return nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("servicebindingrequest-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Register handling of ServiceBindingRequest objects..
	if err = addServiceBindingRequestWatches(c); err != nil {
		return err
	}

	// White list GVKs that should be observed to reconcile an existing related
	// ServiceBindingRequest.
	whiteListedGVKs := []schema.GroupVersionKind{
		{Group: "", Version: "v1", Kind: "Secret"},
	}
	if err = common.AddWatchesWithGVKs(c, whiteListedGVKs, common.ReconcileRelatedSBR); err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileServiceBindingRequest implements reconcile.Reconciler
var _ reconcile.Reconciler = &Reconciler{}
