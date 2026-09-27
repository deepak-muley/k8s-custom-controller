package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestK8sCustomController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8sCustomController Suite")
}
