package main

import (
	"axgle/mahonia"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag 	   int  // 当前client的模式
}

var serverIp string
var serverPort int

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp: serverIp,
		ServerPort: serverPort,
		flag: -1,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}
	client.conn = conn
	return client
}

/*
DealResponse 处理server回应的消息，直接显示到标准输出
 */
func (client *Client) DealResponse() {
	// 一旦client.conn有数据，直接copy到stdout标准输出，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

/*
menu 客户端模式菜单
 */
func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>> 请输入合法范围内的数字 <<<<<")
		return false
	}
}

/*
init 客户端参数设置
 */
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址（默认是127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口（默认是8888）")
}

func (client *Client) PublicChat() {
	// 提示用户输入消息
	var chatMsg string
	fmt.Println(">>>>>> 请输入聊天内容，exit退出: ")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			// 转码gbk
			enc := mahonia.NewEncoder("GBK")
			sendMsg = enc.ConvertString(sendMsg)
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>>> 请输入聊天内容，exit退出: ")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.QueryUser()
	fmt.Println(">>>>>> 请输入聊天对象|用户名|，exit退出: ")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>>> 请输入消息内容，exit退出: ")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				// 转码gbk
				enc := mahonia.NewEncoder("GBK")
				sendMsg = enc.ConvertString(sendMsg)
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>>> 请输入消息内容，exit退出: ")
			fmt.Scanln(&chatMsg)
		}
		remoteName = ""
		client.QueryUser()
		fmt.Println(">>>>>> 请输入聊天对象|用户名|，exit退出: ")
		fmt.Scanln(&remoteName)
	}
}

/*
UpdateName 更新用户名
 */
func (client *Client) UpdateName() bool {
	fmt.Println(">>>>>> 请输入用户名：")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	// 转码gbk
	enc := mahonia.NewEncoder("GBK")
	sendMsg = enc.ConvertString(sendMsg)
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (client *Client) QueryUser() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

/*
Run 客户端等待命令执行
 */
func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
		// 根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			// 公聊
			// fmt.Println("公聊模式选择...")
			client.PublicChat()
		case 2:
			// 私聊
			// fmt.Println("私聊模式选择...")
			client.PrivateChat()
		case 3:
			// 更新用户名
			// fmt.Println("更新用户名选择...")
			client.UpdateName()
		}
	}
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>> 服务器连接失败 <<<<<<")
		return
	}

	// 单独开启一个goroutine去处理server的回执消息，避免Run阻塞
	go client.DealResponse()

	fmt.Println(">>>>>> 服务器连接成功 <<<<<<")

	// 启动客户端业务
	client.Run()
}
