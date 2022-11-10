package main

import (
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
	flag       int
}

func newClient(serverIp string, serverPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	//连接 server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.dial error:", err)
		return nil
	}
	client.conn = conn

	//返回对象
	return client
}

func (client *Client) dealResponse() {
	//一但有消息就拷贝到stdout中，永久阻塞
	io.Copy(os.Stdout, client.conn)

	//for {
	//	buf := make()
	//	client.conn.Read(buf)
	//	fmt.Println(buf)
	//}

}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户")
	fmt.Println("0.退出")
	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>请输入合法的数字>>>>")
		return false
	}
}

func (client *Client) publicChat() {
	var chatMsg string
	fmt.Println(">>>>请输入聊天内容，exit退出")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("connect error:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容，exit退出")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>请输入用户名>>>>")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.write error:", err)
		return false
	}

	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
		switch client.flag {
		case 1:
			//公聊模式
			fmt.Println("公聊模式选择。..")
			client.publicChat()
			//break
		case 2:
			//私聊模式
			fmt.Println("私聊模式选择。..")
			//break
		case 3:
			//改用户名
			fmt.Println("改用户名。..")
			//break
			client.UpdateName()
		}
	}
}

var ServerAddress string
var ServerPort int

//./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&ServerAddress, "ip", "127.0.0.1", "服务端默认ip(127.0.0.1)")
	flag.IntVar(&ServerPort, "port", 8888, "设置服务器端口(8888)")
}
func main() {
	flag.Parse()

	client := newClient(ServerAddress, ServerPort)
	if client == nil {
		fmt.Println(">>>>连接服务器失败>>>>")
		return
	}

	//单独开启一个goroutine处理server回执的消息
	go client.dealResponse()

	fmt.Println(">>>>连接服务器成功>>>>")

	//阻塞 启动客户端业务
	//select {}
	client.Run()
}
