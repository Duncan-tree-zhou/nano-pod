#!/bin/bash

os=$(go env GOOS)
arch=$(go env GOARCH)

sudo curl -sL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv4.5.4/kustomize_v4.5.4_${os}_${arch}.tar.gz | tar xvz -C /usr/local/bin/
export PATH=$PATH:/usr/local/bin
