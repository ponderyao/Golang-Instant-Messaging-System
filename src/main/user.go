package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

/*
NewUser 创建在线用户
 */
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	// 启动 goroutine 听当前 user channel
	go user.ListenMessage()

	return user
}

/*
ListenMessage 监听当前 user channel，一旦有消息就直接发给客户端
 */
func (this *User) ListenMessage() {
	for {
		msg := <- this.C

		content := msg + "\n"

		// 转码gbk
		//enc := mahonia.NewEncoder("GBK")
		//content = enc.ConvertString(content)

		this.conn.Write([]byte(content))
	}
}

/*
Online 用户上线
 */
func (this *User) Online() {
	// 用户上线，将用户加入到server的OnlineMap中（map非线程安全，需要加解锁）
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播当前用户上线的消息
	this.server.BroadCast(this, "已上线")
}

/*
Offline 用户下线
 */
func (this *User) Offline() {
	// 用户下线，将用户从server的OnlineMap中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播当前用户下线的消息
	this.server.BroadCast(this, "下线")
}

/*
SendMessage 给当前 user 的客户端发送消息
 */
func (this *User) SendMessage(msg string) {
	// 转码gbk
	//enc := mahonia.NewEncoder("GBK")
	//content := enc.ConvertString(msg)

	this.conn.Write([]byte(msg))
}

/*
QueryUsers 查询当前在线用户的名单
 */
func (this *User) QueryUsers() {
	this.server.mapLock.Lock()
	for _, user := range this.server.OnlineMap {
		onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
		this.SendMessage(onlineMsg)
	}
	this.server.mapLock.Unlock()
}

/*
Rename 修改用户名：rename|张三
 */
func (this *User) Rename(newName string) {
	// 判断name是否重复
	_, ok := this.server.OnlineMap[newName]
	if ok {
		this.SendMessage("当前用户名被使用\n")
	} else {
		this.server.mapLock.Lock()
		delete(this.server.OnlineMap, this.Name)
		this.server.OnlineMap[newName] = this
		this.server.mapLock.Unlock()

		this.Name = newName
		this.SendMessage("您已经更新用户名：" + this.Name + "\n")
	}
}

/*
PrivateChat 私聊
 */
func (this *User) PrivateChat(remoteUser *User, msg string) {
	remoteUser.SendMessage(msg)
}

/*
DoMessage 用户处理消息
 */
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户的名单
		this.QueryUsers()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 修改用户名
		newName := strings.Split(msg, "|")[1]
		this.Rename(newName)
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 私聊
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMessage("消息格式不正确，请使用 \"to|张三|hello\" 格式。\n")
			return
		}
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMessage("该用户名不存在\n")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMessage("无消息内容，请重发\n")
			return
		}
		this.PrivateChat(remoteUser, content + "\n")
	} else {
		this.server.BroadCast(this, msg)
	}
}