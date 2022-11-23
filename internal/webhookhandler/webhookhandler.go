package webhookhandler

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io,sideEffects=none,admissionReviewVersions=v1
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=list;watch
// +kubebuilder:rbac:groups="nanopod",resources=nanopod,verbs=get;list;watch
// +kubebuilder:rbac:groups="nanopod",resources=nanopacher,verbs=get;list;watch

type WebhookHandler interface {
	admission.Handler
	admission.DecoderInjector
}

type nanoPodPatcher struct {
	client  client.Client
	decoder *admission.Decoder
	logger  logr.Logger
}

func NewHandler(client client.Client, logger logr.Logger) WebhookHandler {
	return &nanoPodPatcher{
		client: client,
		logger: logger,
	}
}

func (n *nanoPodPatcher) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &v1.Pod{}
	err := n.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	//n.client.Get(ctx, types.NamespacedName{Name: req.Namespace, Namespace: ""})

	//在 pod 中修改字段
	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (p *nanoPodPatcher) InjectDecoder(d *admission.Decoder) error {
	p.decoder = d
	return nil
}
