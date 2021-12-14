// +build tools

// Place any runtime dependencies as imports in this file.
// Go modules will be forced to download and install them.
package sample_operator

import (
	_ "github.com/onsi/ginkgo/ginkgo"
	_ "github.com/operator-framework/operator-sdk/cmd/operator-sdk"
	_ "golang.org/x/tools/cmd/goimports"
	_ "k8s.io/code-generator"
)
