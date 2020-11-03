package tag

import (
	"github.com/UnderTreeTech/waterdrop/tools/waterdrop/protoc/protobuf/pkg/extensions/gogoproto"
	// nolint:staticcheck
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

func GetMoreTags(field *descriptor.FieldDescriptorProto) *string {
	if field == nil {
		return nil
	}
	if field.Options != nil {
		v, err := proto.GetExtension(field.Options, gogoproto.E_Moretags)
		if err == nil && v.(*string) != nil {
			return v.(*string)
		}
	}
	return nil
}
