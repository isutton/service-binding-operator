package catchall

import (
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/common"
)

var log = logf.KBLog.WithName("eventhandler").WithName("EnqueueRequestForRelatedSBR")

type EnqueueRequestForRelatedSBR struct{}

func (e EnqueueRequestForRelatedSBR) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	sbrSelector, err := common.GetSBRNamespacedNameFromObject(evt.Object)
	if err != nil {
		log.Error(err, "error on extracting SBR namespaced-name from annotations")
		return
	}

	if sbrSelector == nil {
		return
	}

	q.Add(reconcile.Request{NamespacedName: *sbrSelector})
}

func (e EnqueueRequestForRelatedSBR) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	sbrSelector, err := common.GetSBRNamespacedNameFromObject(evt.ObjectNew)
	if err != nil {
		log.Error(err, "error on extracting SBR namespaced-name from annotations")
		return
	}

	if sbrSelector == nil {
		return
	}

	q.Add(reconcile.Request{NamespacedName: *sbrSelector})
}

func (e EnqueueRequestForRelatedSBR) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	sbrSelector, err := common.GetSBRNamespacedNameFromObject(evt.Object)
	if err != nil {
		log.Error(err, "error on extracting SBR namespaced-name from annotations")
		return
	}

	if sbrSelector == nil {
		return
	}

	q.Add(reconcile.Request{NamespacedName: *sbrSelector})
}

func (e EnqueueRequestForRelatedSBR) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	sbrSelector, err := common.GetSBRNamespacedNameFromObject(evt.Object)
	if err != nil {
		log.Error(err, "error on extracting SBR namespaced-name from annotations")
		return
	}

	if sbrSelector == nil {
		return
	}

	q.Add(reconcile.Request{NamespacedName: *sbrSelector})
}
