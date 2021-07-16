package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	SeverIp   string
	SeverPort int
	Name      string
	Con       net.Conn
	flag      int //当前Client的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		SeverIp:   serverIp,
		SeverPort: serverPort,
		flag:      999,
	}

	// 连接server
	conn, error := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if error != nil {
		fmt.Println("Net Dial error:", error)
	}
	// 获得连接
	client.Con = conn

	return client
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "Set IP address(Default 127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "Set Port(Default 8888)")
}

func main() {
	// 解析flag中的值
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("NewClient error")
		return
	}

	// 单独用一个协程处理回执消息
	go client.DealResponse()

	fmt.Println("NewClient success")
	// 启动客户端的业务
	select {}
}

func (this *Client) Menu(client *Client) bool {
	var flag int

	fmt.Println("1:Public chat mode")
	fmt.Println("2:Private chat mode")
	fmt.Println("3:Update username")
	fmt.Println("0:exit")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("Please enter a number within the legal range")
		return false
	}
}

func Run(client *Client) {
	for client.flag != 0 {
		for client.Menu(client) != true {
			//return
		}

		switch client.flag {
		case 1:
			// 公聊模式
			client.PublicChat()
		case 2:
			// 私聊模式
			client.PrivateChat()
		case 3:
			// 更改用户名
			client.UpdateName()
		}
	}
}

// UpdateName
// 更新用户名
func (client *Client) UpdateName() bool {
	fmt.Println("please enter user name")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.Con.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("Write error:", err)
		return false
	}
	return true
}

// DealResponse
// 处理Server的回执消息
func (client *Client) DealResponse() {
	// 一旦client.conn中有数据，则copy到stdout输出，永久阻塞监听
	io.Copy(os.Stdout, client.Con)
}

// PublicChat
// 公聊不停监听发送消息
func (client *Client) PublicChat() {
	// 提示用户输入消息
	var chatMsg string

	fmt.Println("Please enter chat message")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 发送服务器

		// 消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.Con.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("Write error:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println("Please enter chat message")
		fmt.Scanln(&chatMsg)
	}
}

// PrivateChat
// 私聊功能
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	// 显示所有用户
	// func

	fmt.Println("Please enter a chat user")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println("Please enter chat message")
		fmt.Scanln(&chatMsg)
		if len(chatMsg) != 0 {
			sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
			_, err := client.Con.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("Write error:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println("Please enter chat message")
		fmt.Scanln(&chatMsg)
	}
}
