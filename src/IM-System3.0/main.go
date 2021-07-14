package main

func main() {
	// 开启一个Server
	server := NewServer("127.0.0.1", 8899)
	// 连接启动
	server.Start()
}
