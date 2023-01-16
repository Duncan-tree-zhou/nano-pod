package webhookhandler

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	nanopodv1 "nano-pod-operator/api/v1"
	"nano-pod-operator/internal/patcherfactory"
	"net/http"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"strings"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=ignore,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io,sideEffects=none,admissionReviewVersions=v1
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=list;watch
// +kubebuilder:rbac:groups="nanopod",resources=nanopod,verbs=get;list;watch
// +kubebuilder:rbac:groups="nanopod",resources=nanopacher,verbs=get;list;watch

type WebhookHandler interface {
	admission.Handler
	admission.DecoderInjector
}

type nanoPodWebhookHandler struct {
	client         client.Client
	decoder        *admission.Decoder
	logger         logr.Logger
	patcherFactory patcherfactory.PatcherFactory
}

func NewHandler(client client.Client, logger logr.Logger, patcherFactory *patcherfactory.PatcherFactory) WebhookHandler {
	return &nanoPodWebhookHandler{
		client:         client,
		logger:         logger,
		patcherFactory: *patcherFactory,
	}
}

const (
	LabelNanoPods = "nano-pods"
)

var (
	podStructSchema = strategicpatch.PatchMetaFromStruct{
		T: reflect.TypeOf(v1.Pod{}),
	}
)

func (n *nanoPodWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	n.logger.V(1).Info("======================== NanoPod Handler =========================")
	// check if namespace had been created.
	ns := v1.Namespace{}
	err := n.client.Get(ctx, types.NamespacedName{Name: req.Namespace, Namespace: ""}, &ns)
	if err != nil {
		res := admission.Errored(http.StatusInternalServerError, err)
		res.Allowed = true
		return res
	}

	// decode pod raw data to Utd.
	podInfo, err := n.DecodeToUtd(ctx, req, ns)
	if err != nil {
		res := admission.Errored(http.StatusInternalServerError, err)
		res.Allowed = true
		return res
	}

	// to find the matched pods
	nanoPods, err := n.FindNanoPods(ctx, podInfo, ns)
	if err != nil {
		res := admission.Errored(http.StatusInternalServerError, err)
		res.Allowed = true
		return res
	}

	// patch NanoPods to pod
	patchedRaw, err := n.BatchPatch(ctx, podInfo, nanoPods)
	if err != nil {
		n.logger.Error(err, "failed to patch pod with nano pod....")
		res := admission.Errored(http.StatusInternalServerError, err)
		res.Allowed = true
		return res
	}
	n.logger.V(1).Info("succeed to patch pod with nano pod....", "patchedRaw", string(patchedRaw))
	n.logger.V(1).Info("succeed to patch pod with nano pod....", "req.Object.Raw", string(req.Object.Raw))

	return admission.PatchResponseFromRaw(req.Object.Raw, patchedRaw)
}

func (n *nanoPodWebhookHandler) FindNanoPods(ctx context.Context, podUtd *unstructured.Unstructured, ns v1.Namespace) ([]nanopodv1.NanoPod, error) {
	labels := podUtd.GetLabels()
	n.logger.V(1).Info("get labels.....", "labels", labels)
	var nanoPodsStr string
	var ok bool
	if nanoPodsStr, ok = labels[LabelNanoPods]; !ok {
		n.logger.V(1).Info("nano pods is not set....")
	}

	var nanoPodNames = []string{"default"}

	if len(nanoPodsStr) > 0 {
		nanoPodNames = append(nanoPodNames, strings.Split(nanoPodsStr, ",")...)
	}

	nanoPods, err := n.getMatchedNanoPods(ctx, nanoPodNames, &ns)
	if err != nil {
		n.logger.Error(err, "failed to get matched nano pods....")
		return nil, err
	}
	n.logger.V(1).Info("get matched nanoPods .....", "nanoPods", nanoPods)
	return nanoPods, nil
}

func (n *nanoPodWebhookHandler) DecodeToUtd(_ context.Context, req admission.Request, ns v1.Namespace) (*unstructured.Unstructured, error) {
	pod := v1.Pod{}
	err := n.decoder.Decode(req, &pod)
	if err != nil {
		n.logger.Error(err, "failed to decode req.Object.Raw....")
		return nil, err
	}
	n.logger.V(1).Info("pod decoded....", "podGenName", pod.GetGenerateName())

	// prepared unstructured pod info.
	podMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pod)
	if err != nil {
		n.logger.Error(err, "failed to unstructured pod....", "podName", pod.GetName(), "namespace", ns.Namespace)
		return nil, err
	}
	n.logger.V(1).Info("pod to podMap.....", "podMap", podMap)
	podUtd := &unstructured.Unstructured{Object: podMap}

	return podUtd, err
}

func (n *nanoPodWebhookHandler) getMatchedNanoPods(ctx context.Context, nanoPodNames []string, namespace *v1.Namespace) ([]nanopodv1.NanoPod, error) {
	var nanoPods []nanopodv1.NanoPod

	// add the nano pod defined in pod.metadata.annotation["nano-pods"]
	for _, nanoPodName := range nanoPodNames {
		namespacedName := types.NamespacedName{
			Name:      strings.TrimSpace(nanoPodName),
			Namespace: namespace.Name,
		}
		nanoPod := &nanopodv1.NanoPod{}
		err := n.client.Get(ctx, namespacedName, nanoPod)
		if err != nil {
			n.logger.Error(err, "failed to find nano pods....", "nanoPodName", nanoPodName, "namespace", namespace.Name)
			continue
		}
		nanoPods = append(nanoPods, *nanoPod)
	}
	return nanoPods, nil
}

func (n *nanoPodWebhookHandler) BatchPatch(_ context.Context, podInfo *unstructured.Unstructured, nanoPods []nanopodv1.NanoPod) ([]byte, error) {
	podUnstructured := podInfo.Object
	n.logger.V(1).Info("podUtd before....", "podUtd", podUnstructured)
	for _, np := range nanoPods {
		nanoPodUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&(np.Spec.Template))
		n.logger.V(1).Info("nanoPodTemplateUtd ....", "nanoPodTemplateUtd", nanoPodUnstructured)
		if err != nil {
			n.logger.Error(err, "failed to convert nanoPod.spec.template to unstructured.", "nanoPod name", np.Name)
		}
		podUnstructured, err = n.patcherFactory.GetPatcher(np.Spec.PatchStrategy).Patch(podUnstructured, nanoPodUnstructured)
		if err != nil {
			n.logger.Error(err, "failed to patch nanoPod %s.", "nanoPod name", np.Name)
		}
		n.logger.V(1).Info("podUtd patched with ....", "nanoPod", np.Name, "podUtd", podUnstructured)
	}
	return json.Marshal(podUnstructured)
}

func (p *nanoPodWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	p.decoder = d
	return nil
}
