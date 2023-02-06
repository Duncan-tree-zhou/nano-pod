#!/bin/bash

set -x

os=$(go env GOOS)
arch=$(go env GOARCH)
kubebuilder_exec=kubebuilder_${os}_${arch}

curl -L https://github.com/kubernetes-sigs/kubebuilder/releases/download/v3.6.0/${kubebuilder_exec} --output ${kubebuilder_exec}
sudo mv /tmp/${kubebuilder_exec} /usr/local/kubebuilder/bin/kubebuilder
export PATH=$PATH:/usr/local/kubebuilder/bin

