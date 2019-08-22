package catchall

import (
	"github.com/redhat-developer/service-binding-operator/pkg/controller/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Add controller to the worker. Defines the method that will be called during operator bootstrap.
func Add(mgr manager.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

// add watches to the GVKs that this controller is interested on.
func add(mgr manager.Manager, r *CatchAllReconciler) error {
	c, err := controller.New("catchall-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	if err = common.AddWatchesWithGVKs(c, getGVKs(), common.ReconcileRelatedSBR); err != nil {
		return err
	}

	return nil
}

// newReconciler execute the bootstrap of a new Reconciler object.
func newReconciler(mgr manager.Manager) (*CatchAllReconciler, error) {
	dynClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, err
	}

	return &CatchAllReconciler{
		Client:    mgr.GetClient(),
		DynClient: dynClient,
		Scheme:    mgr.GetScheme(),
	}, nil
}

// getGVKs returns a list of GVKs that this controller will watch for changes.
// TODO: this list should be fetched from K8S API-Server, and later apply a blacklist;
func getGVKs() []schema.GroupVersionKind {
	return []schema.GroupVersionKind{
		// {Group: "apps.openshift.io", Version: "v1alpha1", Kind: "ServiceBindingRequest"},
		{Group: "", Version: "v1", Kind: "Secret"},
	}
}
