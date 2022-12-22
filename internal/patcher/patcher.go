package patcher

import (
	v1 "k8s.io/api/core/v1"
	nanoV1 "nano-pod-operator/api/v1"
)

type Patcher struct {
	nanoPods []nanoV1.NanoPod
}

func NewPatcher(nanoPods []nanoV1.NanoPod) Patcher {
	return Patcher{
		nanoPods: nanoPods,
	}
}

func (p *Patcher) patch(srcPod *v1.Pod) (destPod *v1.Pod) {
	return srcPod
}
