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
#### 指定http请求方法
需要指定请求方法的service或者rpc上方添加，
```
//`method:"post"`
```
表示指定为post方法，默认为get方法，如果service和rpc上方都指定了不同的方法，则优先使用rpc的