package render_test
//
//import (
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//	"github.com/solo-io/autopilot/codegen/model"
//	. "github.com/solo-io/autopilot/codegen/render"
//)
//
//var _ = FDescribe("ChartRenderer", func() {
//	It("do", func() {
//		out, err := RenderChart(model.Chart{
//			Operators: []model.Operator{
//				{
//					Name: "painter",
//					Deployment: model.Deployment{
//						Image: model.Image{
//							Repository: "myapp",
//							Tag:        "v1",
//						},
//					},
//				},
//			},
//			Values: struct {
//				Some string `json:"some"`
//			}{
//				Some: "value",
//			},
//			Data: model.Data{
//				ApiVersion:  "v1",
//				Description: "just a happy little chart",
//				Name:        "Bob Ross",
//				Version:     "0.0.1",
//				Home:        "Solo.io",
//				Icon:        "https://www.biography.com/.image/ar_1:1%2Cc_fill%2Ccs_srgb%2Cg_face%2Cq_auto:good%2Cw_300/MTIwNjA4NjMzOTU5NTgxMTk2/bob-ross-9464216-1-402.jpg",
//				Sources:     []string{"https://www.solo.io"},
//			},
//		})
//		Expect(err).NotTo(HaveOccurred())
//		Expect(out).To(HaveLen(3))
//	})
//})
