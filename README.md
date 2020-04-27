# Go support for Protocol Buffers - Google's data interchange format

[![Build Status](https://travis-ci.org/golang/protobuf.svg?branch=master)](https://travis-ci.org/golang/protobuf)
[![GoDoc](https://godoc.org/github.com/golang/protobuf?status.svg)](https://godoc.org/github.com/golang/protobuf)

Google's data interchange format.
Copyright 2010 The Go Authors.
https://github.com/golang/protobuf

This package and the code it generates requires at least Go 1.9.

This software implements Go bindings for protocol buffers.  For
information about protocol buffers themselves, see
	https://developers.google.com/protocol-buffers/

## 介绍 ##

在原本proto-gen-go的基础上魔改，衍生出proto-gen-lightbrother, 添加了gin插件， 使用gin插件后，会根据proto自动生成对应的gin路由。

## 用法 ##

#### 安装
```
    git clone https://github.com/lightbrotherV/gin-protobuf
    cd gin-protobuf/protoc-gen-lightbrother
    go install
```
#### 使用
```
    protoc --lightbrother_out=plugins=gin:. *.proto
    // 同时生成grpc
    protoc --lightbrother_out=plugins=grpc+gin:. *.proto
```
#### 使用中间件
- 申明

    可以在service或者rpc上方添加，多个中间件用逗号","分割。
    中间件调用时分先后, 先调用service上的中间件，再调用rpc的，申明多个中间件，按照申明先后排序调用。
``` golang
    // `middleware:"auth,cors"`
```

- 注册中间处理函数处理函数

    对于每个中间件，插件会生成对应的注册函数，函数名为Register<中间名>Middleware，对于上方例子，则会生成如下注册函数:
```golang
    func RegisterAuthMiddleware(f gin.HandlerFunc);
    func RegisterCorsMiddleware(f gin.HandlerFunc);
```

#### 返回结果
通过gin框架的上下文gin.Context的Get,Set方法来传递参数。
返回结构为：
```
{
    code: 0,
    message: "",
    data: {}
}
```

data为grpc的返回结构体，code为接口返回码，默认为0，message为错误信息,默认为空字符串。
