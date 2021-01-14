package tag

import (
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"google.golang.org/protobuf/proto"
)

func GetMoreTags(field *descriptor.FieldDescriptorProto) *string {
	if field == nil {
		return nil
	}
	if field.Options != nil {
		v := proto.GetExtension(field.Options, nil)
		if v.(*string) != nil {
			return v.(*string)
		}
	}
	return nil
}
