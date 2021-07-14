package main

import (
	"flag"
	"fmt"
	"net"
)

type Client struct {
	SeverIp   string
	SeverPort int
	Name      string
	Con       net.Conn
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		SeverIp:   serverIp,
		SeverPort: serverPort,
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
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置IP地址(默认为127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置端口号(默认为8888)")
}

func main() {
	// 解析flag中的值
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("NewClient error")
		return
	}
	fmt.Println("NewClient success")
	// 启动客户端的业务
	select {}
}
