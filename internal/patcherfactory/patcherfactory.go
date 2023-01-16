package patcherfactory

import (
	nanopodv1 "nano-pod-operator/api/v1"
	"nano-pod-operator/internal/patcherfactory/patcher"
)

type PatcherFactory struct {
	patchers       map[nanopodv1.PatchStrategy]patcher.Patcher
	defaultPatcher patcher.Patcher
}

type builder struct {
	patchers       map[nanopodv1.PatchStrategy]patcher.Patcher
	defaultPatcher patcher.Patcher
}

func NewBuilder() *builder {
	return &builder{
		patchers:       make(map[nanopodv1.PatchStrategy]patcher.Patcher),
		defaultPatcher: nil,
	}
}

func (b *builder) Register(key nanopodv1.PatchStrategy, p patcher.Patcher) *builder {
	if 0 == len(key) || nil == p {
		panic("key and patcher can not be nil.")
	}
	if nil == b.patchers {
		b.patchers = make(map[nanopodv1.PatchStrategy]patcher.Patcher)
	}

	b.patchers[key] = p

	if nil == b.defaultPatcher {
		b.defaultPatcher = p
	}
	return b
}

func (b *builder) Build() *PatcherFactory {
	if len(b.patchers) <= 0 {
		panic("can not build an empty PatcherFactory.")
	}
	if nil == b.defaultPatcher {
		panic("default patcher can not be nil.")
	}
	return &PatcherFactory{
		patchers:       b.patchers,
		defaultPatcher: b.defaultPatcher,
	}
}

func (pf *PatcherFactory) GetPatcher(patchStrategy nanopodv1.PatchStrategy) patcher.Patcher {
	patcher, ok := pf.patchers[patchStrategy]
	if !ok {
		return pf.defaultPatcher
	} else {
		return patcher
	}
}
