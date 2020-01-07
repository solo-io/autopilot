package render_test
//
//import (
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//	"github.com/solo-io/autopilot/codegen/proto"
//	. "github.com/solo-io/autopilot/codegen/render"
//	"github.com/solo-io/autopilot/codegen/util"
//	"github.com/solo-io/autopilot/codegen/writer"
//	"github.com/solo-io/autopilot/test"
//	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//
//	//"github.com/solo-io/autopilot/test"
//	"k8s.io/apimachinery/pkg/runtime/schema"
//)
//
//var _ = Describe("Generated Clients", func() {
//
//	It("uses the generated clientsets to crud", func() {
//		clientset, err := versioned.NewForConfig(test.MustConfig())
//		Expect(err).NotTo(HaveOccurred())
//
//		things, err := clientset.ThingsV1().Paints("default").List(v1.ListOptions{})
//		Expect(err).NotTo(HaveOccurred())
//		Expect(things.Items).To(HaveLen(1))
//	})
//})
