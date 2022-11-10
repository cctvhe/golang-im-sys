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

//创建一个用户的api
func newUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	//启动监听
	go user.listenMessage()

	return user
}

//用户上线
func (this *User) Online() {
	//用户上线，将用户l加入到onlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播当前用户上线消息
	this.server.broadCast(this, "上线啦")
}

//用户下线
func (this *User) Offline() {
	//用户上线，将用户l加入到onlineMap中
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播当前用户上线消息
	this.server.broadCast(this, "下线啦")
}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

//用户处理消息
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前用户在线都有那些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式: rename|张三
		newName := strings.Split(msg, "|")[1]

		//判断name是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("当前用户名被使用\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("您已经更新用户名" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式 to|张三|消息内容
		remoteName := strings.Split(msg, "|")[1]

		//1.获取对方的用户名
		if remoteName == "" {
			this.SendMsg("消息格式不正确,请使用 \"to|张三|你好啊\"格式\n")
			return
		}

		//2.根据用户名，获取对方的user对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("该用户名不存在\n")
			return
		}

		//3.获取消息内容，根据对方的user对象将消息发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("无消息内容，请重发\n")
			return
		}
		remoteUser.SendMsg(this.Name + "对您说:" + content)
	} else {
		this.server.broadCast(this, msg)
	}
}

//监听当前user channel方法
func (this *User) listenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
