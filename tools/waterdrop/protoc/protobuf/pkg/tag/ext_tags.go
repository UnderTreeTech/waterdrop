package tag

import (
	"google.golang.org/protobuf/proto"
	descriptor "google.golang.org/protobuf/types/descriptorpb"
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
