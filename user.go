package main

import "net"

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

//用户处理消息
func (this *User) DoMessage(msg string) {
	this.server.broadCast(this, msg)
}

//监听当前user channel方法
func (this *User) listenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
