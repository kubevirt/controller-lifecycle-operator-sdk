package sdk

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSdk(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SDK Suite")
}
