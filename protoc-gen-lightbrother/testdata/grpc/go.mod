module github.com/lightbrotherV/gin-protobuf/protoc-gen-go/testdata/grpc

go 1.9

require (
	github.com/golang/protobuf v1.3.4
	google.golang.org/grpc v1.29.1
)

replace github.com/golang/protobuf => ../../..
