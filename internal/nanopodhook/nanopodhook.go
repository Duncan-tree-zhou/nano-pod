package nanopodhook

import (
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
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

type NanoPodHook interface {
	admission.Handler
	admission.DecoderInjector
}

type nanoPodWebhookHandler struct {
	client  client.Client
	decoder *admission.Decoder
	logger  logr.Logger
}

func NewHandler(client client.Client, logger logr.Logger) NanoPodHook {
	return &nanoPodWebhookHandler{
		client: client,
		logger: logger,
	}
}

const (
	AnnotationNanoPods = "nano-pods"
)

var (
	mergeItemStructSchema = strategicpatch.PatchMetaFromStruct{}
)

func (n *nanoPodWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	podInfo := unstructured.Unstructured{}
	err := json.Unmarshal(req.Object.Raw, podInfo)
	if err != nil {
		res := admission.Errored(http.StatusInternalServerError, err)
		res.Allowed = true
		return res
	}

	annotations := podInfo.GetAnnotations()
	var nanoPodsStr string
	var ok bool
	if nanoPodsStr, ok = annotations[AnnotationNanoPods]; !ok {
		return admission.Allowed("no nano pod annotation")
	}

	var nanoPodNames = []string{"default"}
	nanoPodNames = append(nanoPodNames, strings.Split(nanoPodsStr, ",")...)

	nanoPods, err := n.getMatchedNanoPods(ctx, &podInfo, nanoPodNames)
	if err != nil {
		res := admission.Errored(http.StatusInternalServerError, err)
		res.Allowed = true
		return res
	}

	patchedRaw, err := nanoPodsPatch(ctx, &podInfo, nanoPods)
	if err != nil {
		res := admission.Errored(http.StatusInternalServerError, err)
		res.Allowed = true
		return res
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, patchedRaw)
}

func (n *nanoPodWebhookHandler) getMatchedNanoPods(ctx context.Context, podInfo *unstructured.Unstructured, nanoPodNames []string) ([]nanopodv1.NanoPod, error) {
	var nanoPods []nanopodv1.NanoPod

	namespace := podInfo.GetNamespace()

	// add the nano pod defined in pod.metadata.annotation["nano-pods"]
	for _, nanoPodName := range nanoPodNames {
		namespacedName := types.NamespacedName{
			Name:      strings.TrimSpace(nanoPodName),
			Namespace: namespace,
		}
		nanoPod := &nanopodv1.NanoPod{}
		err := n.client.Get(ctx, namespacedName, nanoPod)
		if err != nil {
			nanoPods = append(nanoPods, *nanoPod)
		}
	}
	return nanoPods, nil
}

func nanoPodsPatch(ctx context.Context, podInfo *unstructured.Unstructured, nanoPods []nanopodv1.NanoPod) ([]byte, error) {
	podUnstructured := podInfo.Object
	for _, np := range nanoPods {
		nanoPodUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(np.Spec)
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
