// +build tools

package main

import (
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
	_ "github.com/brancz/locutus"
	_ "github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb"
)
