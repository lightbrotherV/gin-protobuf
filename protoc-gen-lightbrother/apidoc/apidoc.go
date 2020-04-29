package apidoc

import (
	"bytes"
	"fmt"
	"github.com/lightbrotherV/gin-protobuf/protoc-gen-lightbrother/descriptor"
	"github.com/lightbrotherV/gin-protobuf/protoc-gen-lightbrother/generator"
	plugin_go "github.com/lightbrotherV/gin-protobuf/protoc-gen-lightbrother/plugin"
	"path"
	"strconv"
	"strings"
)

func init() {
	generator.RegisterPlugin(new(apidoc))
}

type apidoc struct {
	gen                 *generator.Generator
	serviceMethodRoutes map[string]string
	messageType         map[string]*DescriptorProto
	*bytes.Buffer
}

func (a *apidoc) Name() string {
	return "apidoc"
}

func (a *apidoc) Init(gen *generator.Generator) {
	a.gen = gen
	a.Buffer = new(bytes.Buffer)
	a.serviceMethodRoutes = make(map[string]string)
	a.messageType = make(map[string]*DescriptorProto)
}

func (a *apidoc) Generate(file *generator.FileDescriptor) {
	if len(file.GetService()) == 0 {
		return
	}
	a.Reset()
	a.setMessageType(file)
	name := getFileName(file.GetName())

	a.GenerateTop(file)
	a.GenerateMethod(file)

	//提取生成的代码
	content := a.String()
	ginFile := &plugin_go.CodeGeneratorResponse_File{
		Name:    &name,
		Content: &content,
	}
	resFile := a.gen.Response.File
	resFile = append(resFile, ginFile)
	a.gen.Response.File = resFile
}

func (a *apidoc) GenerateImports(file *generator.FileDescriptor) {

}

func (a *apidoc) GenerateTop(file *generator.FileDescriptor) {
	services := file.GetService()
	packageName := file.GetPackage()
	for si, serv := range services {
		servName := serv.GetName()
		methods := serv.GetMethod()
		for mi, method := range methods {
			methodCommentsPath := fmt.Sprintf("6,%d,2,%d", si, mi)
			comment := file.Comments[methodCommentsPath].GetLeadingComments()
			comments := generator.GetCommentWithoutTag(comment)
			if len(comments) > 0 {
				comment = comments[0]
			} else {
				comment = ""
			}
			methodName := method.GetName()
			target := fmt.Sprintf("%s%s%s", strings.ToLower(strings.ReplaceAll(packageName, ".", "")), strings.ToLower(servName), strings.ToLower(methodName))
			a.P(fmt.Sprintf("- [/%s.%s/%s](#%s) %s", packageName, servName, methodName, target, comment))
		}
	}
}

func (a *apidoc) GenerateMethod(file *generator.FileDescriptor) {
	a.P()
	services := file.GetService()
	packageName := file.GetPackage()
	for si, serv := range services {
		servName := serv.GetName()
		methods := serv.GetMethod()
		for mi, method := range methods {
			methodCommentsPath := fmt.Sprintf("6,%d,2,%d", si, mi)
			comment := file.GetLineComment(methodCommentsPath)
			methodName := method.GetName()
			a.P(fmt.Sprintf("##/%s.%s/%s", packageName, servName, methodName))
			if comment != "" {
				a.P(fmt.Sprintf("###%s", comment))
			}
			a.P()
			a.P("#### 方法：GRPC")
			a.P()
			a.P("#### 请求参数")
			a.P()

			methodInput := getServiceMessageName(method.GetInputType())
			fields := a.messageType[methodInput].GetField()
			result:=&bytes.Buffer{}
			for i, field := range fields {
				// 如果是参数里有message,就只能用json展示了。
				if field.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
					a.P("```javascript")
					a.P(a.exampleJsonForMsg(a.messageType[methodInput], 4))
					a.P("```")
					a.P()
					result = &bytes.Buffer{}
					break
				}
				if i == 0 {
					a.P("|参数名|必选|类型|描述|")
					a.P("|:---|:---|:---|:---|")
				}
				require := "否"
				if isRequired(field.GetLabel()) {
					require = "是"
				}
				result.WriteString(fmt.Sprintf("|%s|%s|%s|%s|\n",field.GetJsonName(), require, getParamType(field.GetType()), a.messageType[methodInput].FelidComment[field.GetName()]))
			}
			a.P(result.String())
			a.P()
			a.P("#### 响应")
			a.P()
			methodOutput := getServiceMessageName(method.GetOutputType())
			OutputJsonStr := a.exampleJsonForMsg(a.messageType[methodOutput], 4)
			a.P("```javascript")
			a.P(`{
    "code": 0,
    "message": "ok",
    "data": `)
			a.P(OutputJsonStr)
			a.P("}")
			a.P("```")
			a.P()
			a.P()
		}
	}
}

func (a *apidoc) exampleJsonForMsg(messageDescriptor *DescriptorProto, indent int) string {
	buf := &bytes.Buffer{}
	buf.WriteString(strings.Repeat(" ", indent))
	buf.WriteString("{\n")
	felids := messageDescriptor.GetField()
	for _, felid := range felids {
		buf.WriteString(fmt.Sprintf("%s//%s\n", strings.Repeat(" ", indent+4),messageDescriptor.FelidComment[felid.GetJsonName()]))
		if isRepeated(felid.GetLabel()) {
			buf.WriteString(fmt.Sprintf("%s\"%s\": [\n", strings.Repeat(" ", indent+4), felid.GetJsonName()))
		} else {
			buf.WriteString(fmt.Sprintf("%s\"%s\": ", strings.Repeat(" ", indent+4), felid.GetJsonName()))
		}

		if felid.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			typeNameArr := strings.Split(felid.GetTypeName(), ".")
			typeName := typeNameArr[len(typeNameArr)-1]
			preIndext := indent + 4
			if isRepeated(felid.GetLabel()) {
				preIndext += 4
			}
			buf.WriteString(a.exampleJsonForMsg(messageDescriptor.SubMessageType[typeName], preIndext))
		} else {
			if isRepeated(felid.GetLabel()) {
				buf.WriteString(fmt.Sprintf("%s%s\n", strings.Repeat(" ", indent+8), getParamTypeDefault(felid.GetType())))
			} else {
				buf.WriteString(fmt.Sprintf("%s\n", getParamTypeDefault(felid.GetType())))
			}
		}

		if isRepeated(felid.GetLabel()) {
			buf.WriteString(fmt.Sprintf("%s]\n", strings.Repeat(" ", indent+4)))
		}
	}
	buf.WriteString(fmt.Sprintf("%s}\n", strings.Repeat(" ", indent)))
	return buf.String()
}

func (a *apidoc) setMessageType(file *generator.FileDescriptor) {
	messageTypesDescriptor := file.GetMessageType()
	for i, messageDescriptor := range messageTypesDescriptor {
		a.messageType[messageDescriptor.GetName()] = NewDescriptorProto(file, messageDescriptor, fmt.Sprintf("4,%d", i))
	}
}

// 写入缓冲区
func (g *apidoc) P(str ...interface{}) {
	for _, v := range str {
		g.printAtom(v)
	}
	g.WriteByte('\n')
}

// printAtom prints the (atomic, non-annotation) argument to the generated output.
func (g *apidoc) printAtom(v interface{}) {
	switch v := v.(type) {
	case string:
		g.WriteString(v)
	case *string:
		g.WriteString(*v)
	case bool:
		fmt.Fprint(g, v)
	case *bool:
		fmt.Fprint(g, *v)
	case int:
		fmt.Fprint(g, v)
	case *int32:
		fmt.Fprint(g, *v)
	case *int64:
		fmt.Fprint(g, *v)
	case float64:
		fmt.Fprint(g, v)
	case *float64:
		fmt.Fprint(g, *v)
	case generator.GoPackageName:
		g.WriteString(string(v))
	case generator.GoImportPath:
		g.WriteString(strconv.Quote(string(v)))
	}
}

// 获取文件名称
func getFileName(name string) string {
	if ext := path.Ext(name); ext == ".proto" || ext == ".protodevel" {
		name = name[:len(name)-len(ext)]
	}
	name += ".md"
	return name
}

// 获取方法，message名称
func getServiceMessageName(type_ string) string {
	typeArr := strings.Split(type_, ".")
	return typeArr[len(typeArr)-1]
}
