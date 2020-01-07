package test_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/autopilot/codegen/proto"
	"github.com/solo-io/autopilot/codegen/util"
	"k8s.io/apimachinery/pkg/runtime/schema"

	. "github.com/solo-io/autopilot/codegen"
)
/*
protoc --gogo_out=plugins=grpc,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor,Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/rpc/status.proto=github.com/gogo/googleapis/google/rpc,Menvoy/api/v2/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2,gogoproto/gogo.proto=github.com/gogo/protobuf/gogoproto:/var/folders/yc/c2kl5bhn5nlfl1518vq_nf4r0000gn/T/724752984 --ext_out=plugins=grpc,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/de -o/var/folders/yc/c2kl5bhn5nlfl1518vq_nf4r0000gn/T/solo-kit-gen-894996951 --include_imports --include_source_info /Users/ilackarms/go/src/github.com/solo-io/autopilot/codegen/test/test_api.proto


 */
var _ = Describe("Render", func() {
	It("compiles protos", func() {
		err := proto.CompileProtos(".")
		Expect(err).NotTo(HaveOccurred())
	})
	It("renders the files for the group", func() {
		files, err := RenderGroup(Group{
			GroupVersion: schema.GroupVersion{
				Group:   "things.test.io",
				Version: "v1",
			},
		Resources: []Resource{
			{
				Kind:   "Paint",
				Spec:   Field{Type: "PaintColor"},
				Status: &Field{Type: "TubeStatus"},
			},
		},
		})
		Expect(err).NotTo(HaveOccurred())

		w := Writer{
			Root:           "api",
			ForceOverwrite: false,
		}

		err = w.WriteFiles(files)
		Expect(err).NotTo(HaveOccurred())

		err = util.DeepcopyGen("./api/things.test.io/v1")
		Expect(err).NotTo(HaveOccurred())
	})
})
