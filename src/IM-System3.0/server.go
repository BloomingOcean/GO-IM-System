package main

import (
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
	OnLineMap map[string]*User
	MapLock   sync.RWMutex
	// 消息广播的channel
	// 用于接受用户上线后发送给server的数据
	// 然后server再把这个数据广播给所有在线(OnLineMap)用户
	Message chan string
}

// NewServer
// 新建一个socket连接
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnLineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// Start
// 启动通信
func (this *Server) Start() {
	// socket listen
	listener, error := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if error != nil {
		fmt.Println("NewServer error:", error)
		return
	}

	// 监听Server的channel
	go this.ListenMessage()

	for {
		// accept
		conn, error := listener.Accept()
		if error != nil {
			fmt.Println("Accept error:", error)
			continue
		}
		// do handler
		go this.Handler(conn)
	}

	// close
	defer listener.Close()
}

// ListenMessage
// 广播到每个用户
// 负责把Server中channel的数据发送到各个用户的channel中
func (this *Server) ListenMessage() {
	for {
		message := <-this.Message
		// 将这条信息发送给所有在线用户
		this.MapLock.Lock()
		for _, user := range this.OnLineMap {
			user.C <- message
		}
		this.MapLock.Unlock()
	}
}

// Handler
// 建立连接之后执行的操作
func (this *Server) Handler(conn net.Conn) {
	fmt.Println("connect success...")

	// 建立上线用户的信息
	user := NewUser(conn, this)

	// 上线功能
	user.OnLine()

	// 用户在线的凭证
	isAlive := make(chan bool)

	// 接受客户端发送的消息(匿名方法)
	go func() {
		bytes := make([]byte, 4096)
		for {
			n, err := conn.Read(bytes)
			// EOF 是当没有更多输入可用时 Read 返回的错误
			if err != nil && err != io.EOF {
				fmt.Print("Read error:", err)
				return
			}
			if n == 0 {
				// 用户下线
				user.OffLine()
				return
			}

			// 切割string的最后的 '\n'
			// msg := string(bytes[:n-1])
			msg := string(bytes[:n])

			// 广播用户发送的消息
			user.DoMessage(msg)
			// 标志用户存活
			isAlive <- true
		}
	}()

	// 当前handler阻塞
	for {
		select {
		case <-isAlive:
			// 当前用户是否活跃,应该重置定时器
			// 不用做操作,select会把每个case的条件执行
		case <-time.After(time.Second * 30):
			// 已经超时
			// 将当前的user强制关闭
			user.SendMessage("You got kicked")
			close(user.C)
			conn.Close()
			// 推出当前Handler
			return
		}
	}
}

// BroadCast
// 设置广播消息信息
func (this *Server) BroadCast(user *User, msg string) {
	this.Message <- fmt.Sprintf("User %s with ip address %s say: %s", user.Name, user.Addr, msg)
}
