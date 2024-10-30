package clock_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const threshold = 5 * time.Millisecond
const mockRatio = 2
const mockThreshold = mockRatio * threshold

func TestUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "clock Suite")
}
