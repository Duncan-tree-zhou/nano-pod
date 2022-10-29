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

func (a *nanoPodPatcher) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &v1.Pod{}
	err := a.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

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
