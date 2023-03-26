package patcher

import (
	"nano-pod-operator/internal/patcherfactory/patcher/overwritepatcher"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

var (
	podStructSchema = strategicpatch.PatchMetaFromStruct{
		T: reflect.TypeOf(v1.Pod{}),
	}
)

type Patcher interface {
	Patch(original map[string]interface{}, patch map[string]interface{}) (map[string]interface{}, error)
}

type OverWritePatcher struct {
}

func (sm *OverWritePatcher) Patch(original map[string]interface{}, patch map[string]interface{}) (map[string]interface{}, error) {
	meta, err := strategicpatch.StrategicMergeMapPatchUsingLookupPatchMeta(original, patch, podStructSchema)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

type DefaultPatcher struct {
}

func (dp *DefaultPatcher) Patch(original map[string]interface{}, patch map[string]interface{}) (map[string]interface{}, error) {
	meta, err := overwritepatcher.Patch(original, patch)
	if err != nil {
		return nil, err
	}
	return meta, nil
}
