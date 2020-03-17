package gin

import (
	"bytes"
	"fmt"
	pb "github.com/golang/protobuf/protoc-gen-lightbrother/descriptor"
	"github.com/golang/protobuf/protoc-gen-lightbrother/generator"
	plugin_go "github.com/golang/protobuf/protoc-gen-lightbrother/plugin"
	"path"
	"strconv"
	"strings"
)

const (
	ginPkgPath     = "github.com/gin-gonic/gin"
	contextPkgPath = "context"
	httpPkgPath    = "net/http"
)

func init() {
	generator.RegisterPlugin(new(gin))
}

type gin struct {
	gen                 *generator.Generator
	serviceMethodRoutes map[string]string
	*bytes.Buffer
}

func (g *gin) Name() string {
	return "gin"
}

func (g *gin) Init(gen *generator.Generator) {
	g.gen = gen
	g.Buffer = new(bytes.Buffer)
	g.serviceMethodRoutes = make(map[string]string)
}

func (g *gin) Generate(file *generator.FileDescriptor) {
	if len(file.GetService()) == 0 {
		return
	}
	g.Reset()
	g.generateHeader(file)
	name := getFileName(file.GetName())
	//json, _ := json.Marshal(file)
	//content := fmt.Sprint(string(json))
	g.generateServiceRoute(file)
	g.generateService(file)
	//提取生成的代码
	content := g.String()
	ginFile := &plugin_go.CodeGeneratorResponse_File{
		Name:    &name,
		Content: &content,
	}
	resFile := g.gen.Response.File
	resFile = append(resFile, ginFile)
	g.gen.Response.File = resFile
}

func (g *gin) GenerateImports(file *generator.FileDescriptor) {

}

func (g *gin) generateHeader(file *generator.FileDescriptor) {
	goPackageName := string(file.PackageName)
	g.P("// Code generated by protoc-gen-lightbrother, DO NOT EDIT.")
	g.P()
	g.P("/*")
	g.P(fmt.Sprintf("Package %s is a generated gin stub package.", goPackageName))
	g.P("This code was generated with protoc-gen-lightbrother. ")
	g.P()
	g.P("It is generated from these files:")
	g.P(fmt.Sprintf("\t%s", file.GetName()))
	g.P("*/")
	g.P(fmt.Sprintf("package %s", goPackageName))
	g.P()
	g.P("import (")
	g.P(fmt.Sprintf("\t%s", strconv.Quote(ginPkgPath)))
	g.P(fmt.Sprintf("\t%s", strconv.Quote(contextPkgPath)))
	g.P(fmt.Sprintf("\t%s", strconv.Quote(httpPkgPath)))
	g.P(")")
	g.P()
	g.P("// to suppressed 'imported but not used warning'")
	g.P()
}

func (g *gin) generateService(file *generator.FileDescriptor) {
	services := file.GetService()
	packageName := file.PackageName
	for i, serv := range services {
		servName := generator.CamelCase(serv.GetName())
		servCommentsPath := fmt.Sprintf("6,%d", i)
		g.P(fmt.Sprintf("// %sGinServer is the server API for %s service.", servName, servName))
		comments := file.Comments[servCommentsPath].GetLeadingComments()
		g.printComments(comments, "")
		g.setServiceMethodRoutes(serv.GetName(), comments)
		g.P(fmt.Sprintf("type %sGinServer interface {", servName))
		g.generateInterfaceProperties(file, serv, i)
		g.P("}")
		g.P()
		g.P(fmt.Sprintf("var %s%sSvc %sGinServer", packageName, servName, servName))
		g.P("var ctx = context.Background()")
		g.P()
		g.generateHandleFunc(file, serv)
		g.generateRegister(file, serv, i)
	}
}

func (g *gin) generateServiceRoute(file *generator.FileDescriptor) {
	packageName := file.GetPackage()
	services := file.GetService()
	for _, serv := range services {
		originServName := serv.GetName()
		for _, method := range serv.Method {
			g.P(fmt.Sprintf("var %s = \"/%s.%s/%s\"", g.getRouteVariable(serv, method), packageName, originServName, method.GetName()))
		}
	}
	g.P()
}

func (g *gin) generateInterfaceProperties(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	methods := service.GetMethod()
	serviceName := service.GetName()
	for i, method := range methods {
		methodCommentsPath := fmt.Sprintf("6,%d,2,%d", index, i)
		comments := file.Comments[methodCommentsPath].GetLeadingComments()
		g.printComments(comments, "\t")
		g.setServiceMethodRoutes(fmt.Sprintf("%s:%s", serviceName, method.GetName()), comments)
		methodName := generator.CamelCase(method.GetName())
		g.P(fmt.Sprintf("\t%s(ctx context.Context, req *%s) (resp *%s, err error)", methodName, g.gen.TypeName(g.objectNamed(method.GetInputType())), g.gen.TypeName(g.objectNamed(method.GetOutputType()))))
		g.P()
	}
}

func (g *gin) generateHandleFunc(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto) {
	methods := service.GetMethod()
	packageName := file.PackageName
	servName := generator.CamelCase(service.GetName())
	for _, method := range methods {
		g.P(fmt.Sprintf("func %s(c *gin.Context) {", method.GetName()))
		g.P(fmt.Sprintf("\tp := new(%s)", g.gen.TypeName(g.objectNamed(method.GetOutputType()))))
		g.P("\tif err := c.BindJSON(p); err!= nil {")
		g.P("\t\tc.JSON(http.StatusInternalServerError, err)")
		g.P("\t}")
		g.P(fmt.Sprintf("\tresp, err := %s%sSvc.%s(ctx, p)", packageName, servName, generator.CamelCase(method.GetName())))
		g.P("\tif err!= nil {")
		g.P("\t\tc.JSON(http.StatusInternalServerError, err)")
		g.P("\t}")
		g.P("\tc.JSON(http.StatusOK, resp)")
		g.P("}")
		g.P()
	}
}

func (g *gin) generateRegister(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	originServiceName := service.GetName()
	servName := generator.CamelCase(originServiceName)
	methods := service.GetMethod()
	packageName := file.PackageName
	g.P(fmt.Sprintf("func Register%sGinServer(e *gin.Engine, server %sGinServer) {", servName, servName))
	g.P(fmt.Sprintf("\t%s%sSvc = server", packageName, servName))
	for _, method := range methods {
		routeMethod := g.getMethodRoute(originServiceName, method.GetName())
		g.P(fmt.Sprintf("\te.%s(%s, %s)", routeMethod, g.getRouteVariable(service, method), method.GetName()))
	}
	g.P("}")
}

// 获取路由变量名
func (g *gin) getRouteVariable(serv *pb.ServiceDescriptorProto, method *pb.MethodDescriptorProto) string {
	servName := generator.CamelCase(serv.GetName())
	methodName := generator.CamelCase(method.GetName())
	return fmt.Sprintf("Path%s%s", servName, methodName)
}

// 写入缓冲区
func (g *gin) P(str ...interface{}) {
	for _, v := range str {
		g.printAtom(v)
	}
	g.WriteByte('\n')
}

// 打印注释
func (g *gin) printComments(comment string, pre string) {
	comments := GetCommentWithoutTag(comment)
	for _, line := range comments {
		g.P(fmt.Sprintf("%s//%s", pre, line))
	}
}

// printAtom prints the (atomic, non-annotation) argument to the generated output.
func (g *gin) printAtom(v interface{}) {
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
	name += ".gin.go"
	return name
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (g *gin) objectNamed(name string) generator.Object {
	return g.gen.ObjectNamed(name)
}

// 从注释中提取记录服务http请求方法
func (g *gin) setServiceMethodRoutes(serviceName, comments string) {
	tags := GetTagsInComment(comments)
	method := GetTagValue("method", tags)
	if method != "" {
		g.serviceMethodRoutes[serviceName] = method
	}
}

// 获取http的请求方式
func (g *gin) getMethodRoute(serviceName, methodName string) string {
	methodKey := fmt.Sprintf("%s:%s", serviceName, methodName)
	if routeMethod, ok := g.serviceMethodRoutes[methodKey]; ok {
		return strings.ToUpper(routeMethod)
	}
	if routeMethod, ok := g.serviceMethodRoutes[serviceName]; ok {
		return strings.ToUpper(routeMethod)
	}
	return "GET"
}
