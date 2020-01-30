// This file contains generated Deepcopy methods for
// Protobuf types with oneofs

package v1

import (
	fmt "fmt"
	math "math"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *PaintSpec) DeepCopyInto(out *PaintSpec) {
	p := proto.Clone(in).(*PaintSpec)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *OilType) DeepCopyInto(out *OilType) {
	p := proto.Clone(in).(*OilType)
	*out = *p
}

// DeepCopyInto supports using AttributeManifest within kubernetes types, where deepcopy-gen is used.
func (in *PaintStatus) DeepCopyInto(out *PaintStatus) {
	p := proto.Clone(in).(*PaintStatus)
	*out = *p
}
