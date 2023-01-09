package patcherfactory

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	nanopodv1 "nano-pod-operator/api/v1"
	"nano-pod-operator/internal/patcherfactory/patcher"
	"reflect"
)

type PatcherFactory struct {
	Patchers map[nanopodv1.PatchStrategy]patcher.Patcher
}

var (
	podStructSchema = strategicpatch.PatchMetaFromStruct{
		T: reflect.TypeOf(v1.Pod{}),
	}
	defaultPatcher = &patcher.OverWritePatcher{}
	patcherFactory = PatcherFactory{
		Patchers: map[nanopodv1.PatchStrategy]patcher.Patcher{
			nanopodv1.OverWritePatch: defaultPatcher,
		},
	}
)

func GetPatcher(patchStrategy nanopodv1.PatchStrategy) patcher.Patcher {
	patcher, ok := patcherFactory.Patchers[patchStrategy]
	if !ok {
		return defaultPatcher
	} else {
		return patcher
	}
}
