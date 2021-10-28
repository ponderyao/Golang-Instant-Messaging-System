# Golang-Instant-Messaging-System

基于Go语言实现的即时通信系统，具备简单的于终端运行的服务端与客户端功能

## 使用流程
1. 新建一个终端，编译服务端并运行
```
$ go build -o server main.go server.go user.go
$ ./server
```
2. 新建另一个终端，编译客户端
```
$ go build -o client client.go
```
3. 开启至少三个终端并运行客户端进行测试
```
$ ./client
```
