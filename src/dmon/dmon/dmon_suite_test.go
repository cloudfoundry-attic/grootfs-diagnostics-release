package dmon_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDmon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dmon Suite")
}
