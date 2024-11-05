package clock_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	ms = time.Millisecond
	th = 5 * ms
)

func TestUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "clock Suite")
}
