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
	"k8s.io/klog/v2"
	nanopodv1 "nano-pod-operator/api/v1"
	"net/http"
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
	client  client.Client
	decoder *admission.Decoder
	logger  logr.Logger
}

func NewHandler(client client.Client, logger logr.Logger) WebhookHandler {
	return &nanoPodWebhookHandler{
		client: client,
		logger: logger,
	}
}

const (
	LabelNanoPods = "nano-pods"
)

var (
	mergeItemStructSchema = strategicpatch.PatchMetaFromStruct{}
)

func (n *nanoPodWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	n.logger.Info("nanopod ===========> into handler.....")
	pod := v1.Pod{}
	err := n.decoder.Decode(req, &pod)
	if err != nil {
		n.logger.Error(err, "nanopod ===========> failed to decode req.Object.Raw....")
		res := admission.Errored(http.StatusInternalServerError, err)
		res.Allowed = true
		return res
	}
	//n.logger.Info("nanopod ===========> pod....", "podGenName", pod.GetGenerateName())

	ns := v1.Namespace{}
	err = n.client.Get(ctx, types.NamespacedName{Name: req.Namespace, Namespace: ""}, &ns)
	if err != nil {
		res := admission.Errored(http.StatusInternalServerError, err)
		res.Allowed = true
		return res
	}

	podUtd, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pod)
	if err != nil {
		n.logger.Error(err, "nanopod ===========> failed to unstructured pod....", "podName", pod.GetName(), "namespace", ns.Namespace)
		res := admission.Errored(http.StatusInternalServerError, err)
		res.Allowed = true
		return res
	}
	podInfo := &unstructured.Unstructured{Object: podUtd}
	//n.logger.Info("nanopod ===========> unmarshal pod.....", "genName", podInfo.GetGenerateName())
	//n.logger.Info("nanopod ===========>  podUtd.....", "podUtd", podUtd)

	labels := podInfo.GetLabels()
	n.logger.Info("nanopod ===========> ", "labels", labels)
	var nanoPodsStr string
	var ok bool
	if nanoPodsStr, ok = labels[LabelNanoPods]; !ok {
		n.logger.Info("nanopod ===========> nano pods is not set....")
	}

	var nanoPodNames = []string{"default"}

	if len(nanoPodsStr) > 0 {
		nanoPodNames = append(nanoPodNames, strings.Split(nanoPodsStr, ",")...)
	}

	nanoPods, err := n.getMatchedNanoPods(ctx, nanoPodNames, &ns)
	if err != nil {
		n.logger.Error(err, "nanopod ===========> failed to get matched nano pods....")
		res := admission.Errored(http.StatusInternalServerError, err)
		res.Allowed = true
		return res
	}
	n.logger.Info("nanopod ===========>  nanoPods .....", "nanoPods", nanoPods)

	patchedRaw, err := n.nanoPodsPatch(ctx, podInfo, nanoPods)
	if err != nil {
		n.logger.Error(err, "nanopod ===========> failed to patch pod with nano pod....")
		res := admission.Errored(http.StatusInternalServerError, err)
		res.Allowed = true
		return res
	}

	n.logger.Info("nanopod ===========> succeed to patch pod with nano pod....", "patchedRaw", string(patchedRaw))

	return admission.PatchResponseFromRaw(req.Object.Raw, patchedRaw)
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
			n.logger.Error(err, "nanopod ===========> failed to find nano pods....", "nanoPodName", nanoPodName, "namespace", namespace.Name)
		}
		nanoPods = append(nanoPods, *nanoPod)
	}
	return nanoPods, nil
}

func (n *nanoPodWebhookHandler) nanoPodsPatch(ctx context.Context, podInfo *unstructured.Unstructured, nanoPods []nanopodv1.NanoPod) ([]byte, error) {
	podUnstructured := podInfo.Object
	for _, np := range nanoPods {
		n.logger.Info("nanopod ===========> nanoPod....", "nanoPod", np)
		nanoPodUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&(np.Spec.Template))
		n.logger.Info("nanopod ===========> nanoPodUtd....", "nanoPodUtd", nanoPodUnstructured)
		n.logger.Info("nanopod ===========> podUtd....", "podUtd", podUnstructured)
		if err != nil {
			klog.Infof("failed to patch.", err)
		}
		podUnstructured = nanoPodPatch(podUnstructured, nanoPodUnstructured)
	}
	return json.Marshal(podUnstructured)
}

func nanoPodPatch(original map[string]interface{}, patch map[string]interface{}) map[string]interface{} {
	meta, err := strategicpatch.StrategicMergeMapPatchUsingLookupPatchMeta(original, patch, mergeItemStructSchema)
	if err != nil {
		klog.Infof("failed to patch.", err)
	}
	return meta
}

func (p *nanoPodWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	p.decoder = d
	return nil
}
