package codegen_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/autopilot/codegen/model"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
	"k8s.io/apimachinery/pkg/runtime/schema"

	. "github.com/solo-io/autopilot/codegen"
)

var _ = Describe("Cmd", func() {
	It("generates controller code and manifests for a proto file", func() {

		cmd := &Command{
			Groups: []Group{
				{
					GroupVersion: schema.GroupVersion{
						Group:   "things.test.io",
						Version: "v1",
					},
					Module: "github.com/solo-io/autopilot",
					Resources: []Resource{
						{
							Kind:   "Paint",
							Spec:   Field{Type: Type{Name: "PaintSpec"}},
							Status: &Field{Type: Type{Name: "PaintStatus"}},
						},
					},
					RenderProtos:     true,
					RenderManifests:  true,
					RenderTypes:      true,
					RenderClients:    true,
					RenderController: true,
					ApiRoot:          "codegen/test/api",
				},
			},
			AnyVendorConfig: &sk_anyvendor.Imports{
				Local: []string{"codegen/test/*.proto"},
			},

			Chart: &Chart{
				Operators: []Operator{
					{
						Name: "painter",
						Deployment: Deployment{
							Image: Image{
								Tag:        "v0.0.0",
								Repository: "painter",
								Registry:   "quay.io/solo-io",
								PullPolicy: "IfNotPresent",
							},
						},
						Args: []string{"foo"},
					},
				},
				Values: nil,
				Data: Data{
					ApiVersion:  "v1",
					Description: "",
					Name:        "Painting Operator",
					Version:     "v0.0.1",
					Home:        "https://docs.solo.io/autopilot/latest",
					Sources: []string{
						"https://github.com/solo-io/autopilot",
					},
				},
			},

			ManifestRoot: "codegen/test/chart",
		}

		err := cmd.Execute()
		Expect(err).NotTo(HaveOccurred())
	})
})
