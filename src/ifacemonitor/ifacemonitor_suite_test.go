package ifacemonitor

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/projectcalico/libcalico-go/lib/testutils"
)

func init() {
	testutils.HookLogrusForGinkgo()
}

func TestIfacemonitor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ifacemonitor Suite")
}
