package patcher

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"reflect"
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
