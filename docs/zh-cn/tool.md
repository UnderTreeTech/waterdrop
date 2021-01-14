# waterdrop工具

Clone [waterdrop](https://github.com/UnderTreeTech/waterdrop) 代码，进入tools目录，执行go install

## 依赖安装

安装protoc：官方下载 [protobuf](https://github.com/protocolbuffers/protobuf/releases) 对应平台的二进制包，解压后按照README进行安装即可

Notice：waterdrop采用 [gogo protobuf](https://github.com/golang/protobuf) 做为pb生成工具

安装protoc-gen-go：执行命令 go get -u github.com/golang/protobuf/protoc-gen-go

## pb代码生成器

执行 `waterdrop protoc --grpc your.proto` 即可生成pb代码 

## swagger接口文档

执行 `waterdrop protoc --swagger your.proto` 即可生成对应的swagger api文档

执行 `waterdrop swagger serve your.swagger.json` 可即时预览api定义

## 单元测试UT

执行 `waterdrop utgen your.go` 即可生成该文件下所有方法的单元测试
执行 `waterdrop utgen --func XXXX your.go` 即可生成对应文件下某个方法的单元测试

## 说明

以上操作第一次执行时会相对比较耗时，需要下载各工具相应的依赖包