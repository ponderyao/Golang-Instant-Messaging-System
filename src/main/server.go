package main

import (
	"axgle/mahonia"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex  // 锁

	// 消息广播的channel
	Message chan string
}

/*
NewServer 创建一个Server
 */
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip: ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}
	return server
}

/*
BroadCast 广播消息
 */
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg

	this.Message <- sendMsg
}

/*
ListenMessage 监听 Message 广播消息 channel 的 goroutine，一旦有消息就发送给全部的在线 user
 */
func (this *Server) ListenMessage() {
	for {
		msg := <- this.Message

		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

/*
Handler 处理当前新连接的业务
 */
func (this *Server) Handler(conn net.Conn) {
	// fmt.Println("连接建立成功")

	user := NewUser(conn, this)
	user.Online()

	// 表示用户活跃的channel
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn.Read err:", err)
				return
			}
			msg := string(buf[:n-1]) // 去除'\n'

			// 转码gbk
			enc := mahonia.NewDecoder("GBK")
			msg = enc.ConvertString(msg)

			user.DoMessage(msg)

			// 激活
			isLive <- true
		}
	}()

	// 当前handler阻塞
	for {
		select {
		case <- isLive:
			// 不执行任何业务，仅代表刷新下面的定时器
		case <- time.After(time.Second * 60):
			// 规定60秒没任何操作便超时，强制关闭当前user的连接
			user.SendMessage("你被踢了")
			close(user.C)
			conn.Close()
			return
		}
	}
}

/*
Start 启动服务器并开始监听连接
 */
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
	}
	// close listen socket
	defer listener.Close()

	// 启动监听 Message 的 goroutine
	go this.ListenMessage()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}
		// do handler
		go this.Handler(conn)
	}

}