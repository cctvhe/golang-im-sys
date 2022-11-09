package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播
	message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		message:   make(chan string),
	}
	return server
}

//坚挺在线channel的goroutine,一旦有消息就发送给全部的user
func (this *Server) ListenMessage() {
	for {
		msg := <-this.message

		//将msg发送给全部用户
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//广播消息
func (this *Server) broadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//fmt.Println("建立连接")

	user := newUser(conn, this)

	user.Online()

	//接收客户端发来的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn read err:", err)
				return
			}
			//读取用户消息去除\n
			msg := string(buf[:n-1])

			//广播
			user.DoMessage(msg)
		}
	}()

	//当前handler阻塞
	select {}
}

func (this *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.listen err:", err)
		return
	}

	defer listener.Close()

	//启动坚挺message的goroutine
	go this.ListenMessage()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		//do handler
		go this.Handler(conn)
	}
}
