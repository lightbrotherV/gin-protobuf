package apidoc

import (
	"fmt"
	"github.com/lightbrotherV/gin-protobuf/protoc-gen-lightbrother/descriptor"
	"github.com/lightbrotherV/gin-protobuf/protoc-gen-lightbrother/generator"
)

type DescriptorProto struct {
	file *generator.FileDescriptor
	*descriptor.DescriptorProto
	FelidComment   map[string]string
	SubMessageType map[string]*DescriptorProto
	path           string
	Comment        string
}

func NewDescriptorProto(f *generator.FileDescriptor, d *descriptor.DescriptorProto, p string) *DescriptorProto {
	result := &DescriptorProto{
		file:            f,
		FelidComment:    make(map[string]string),
		SubMessageType:  make(map[string]*DescriptorProto),
		DescriptorProto: d,
		path:            p,
		Comment:         f.GetLineComment(p),
	}
	felids := result.GetField()
	for i, felid := range felids {
		path := fmt.Sprintf("%s,2,%d", p, i)
		result.FelidComment[felid.GetName()] = f.GetLineComment(path)
	}
	result.addSubMessage()
	return result
}

func (d *DescriptorProto) addSubMessage() {
	subDescriptorProtos := d.GetNestedType()
	for i, subDescriptorProto := range subDescriptorProtos {
		subDescriptorProtoObj := NewDescriptorProto(d.file, subDescriptorProto, fmt.Sprintf("%s,3,%d", d.path, i))
		d.SubMessageType[subDescriptorProto.GetName()] = subDescriptorProtoObj
	}
}
