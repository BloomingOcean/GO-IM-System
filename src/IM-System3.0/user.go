package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	con    net.Conn
	server *Server
}

// NewUser
// 创建一个用户
func NewUser(con net.Conn, server *Server) *User {
	userAddr := con.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		con:    con,
		server: server,
	}

	// 启动监听当前user channel消息的goroutine协程
	go user.ListenMessage()

	return user
}

// ListenMessage
// 监听当前user channel的信息(阻塞监听),如果有数据,则输入到server
func (this *User) ListenMessage() {
	for {
		// 依稀记得 <-后面不能有空格(可以有空格)
		msg := <-this.C
		_, err := this.con.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println("Write error:", err)
		}
	}
}

func (this *User) OnLine() {
	// 把上线用户添加进Server中的在线用户(OnLineMap)中
	this.server.MapLock.Lock()
	this.server.OnLineMap[this.Name] = this
	this.server.MapLock.Unlock()

	// 通知Server的channel，表示已获取连接(把用户信息发送到Server的channel)
	// Server的channel阻塞获取数据，一旦获取到了数据，就把数据广播给所有用户的channel
	this.server.BroadCast(this, "online")
}

func (this *User) OffLine() {
	// 把下线用户从在线用户(OnLineMap)中删掉
	this.server.MapLock.Lock()
	delete(this.server.OnLineMap, this.Name)
	this.server.MapLock.Unlock()

	// 通知Server的channel，表示已获取连接(把用户信息发送到Server的channel)
	// Server的channel阻塞获取数据，一旦获取到了数据，就把数据广播给所有用户的channel
	this.server.BroadCast(this, "offline")
}

// DoMessage
// 处理用户请求
func (this *User) DoMessage(msg string) {
	// 查询当前在线的用户有哪些
	//if msg == "who" {
	if msg == "w" {
		this.server.MapLock.Lock()
		for _, user := range this.server.OnLineMap {
			this.SendMessage(user.Name + " is online\n")
		}
		this.server.MapLock.Unlock()
	}
	if msg[:7] == "modify_" {
		// 更新用户名
		//this.Name = msg[7:]
		newName := strings.Split(msg, "_")[1]
		_, ok := this.server.OnLineMap[newName]
		if ok != false {
			this.SendMessage("The name is repeated")
		} else {
			this.server.MapLock.Lock()
			// 没有重复名称,则修改用户名称(会修改当前user以及OnLineMap)
			delete(this.server.OnLineMap, this.Name)
			this.Name = newName
			this.server.OnLineMap[newName] = this
			this.server.MapLock.Unlock()
			this.SendMessage("Successfully modify your username\n")
		}
	}
	// 私聊消息格式为 to|张三|message
	if msg[:3] == "to|" {
		message := strings.Split(msg, "|")
		name := message[1]
		if name == "" {
			this.SendMessage("Your message is in the wrong format")
			return
		}
		user, ok := this.server.OnLineMap[name]
		if !ok {
			this.SendMessage("Do not have this member")
			return
		}
		if message[2] == "" {
			this.SendMessage("Your message is in the wrong format")
			return
		}
		user.SendMessage(this.Name + "tell you:" + message[2])
	}
	this.server.BroadCast(this, msg)
}

// SendMessage
// 发送数据给客户端显示
func (this *User) SendMessage(msg string) {
	this.con.Write([]byte(msg))
}
