package client

import (
	"encoding/json"
	"fmt"
	"lib"
	"net"
	"strings"
	"time"
)

var clientToken string

func OnStartClient(port string) {
	//创建TCP监听地址
	addr, err := net.ResolveTCPAddr("tcp", port)
	lib.CheckErr(err)

	//建立连接
	conn, err := net.DialTCP("tcp", nil, addr)
	lib.CheckErr(err)

	//首次连接时读取服务器端的提示信息
	data := make([]byte, 1024)
	conn.Read(data)
	fmt.Println(string(data))
	//ping
	go checkOnline(conn)

	//输入账户名
	fmt.Print("请输入账户名:")
	username := lib.ScanLine()

	//输入密码
	fmt.Print("请输入密码:")
	password := lib.ScanLine()

	//登录信息
	var loginInfo struct {
		Username string
		Password string
	}

	username = lib.FmtStr(username)
	password = lib.FmtStr(password)

	loginInfo.Username = username
	loginInfo.Password = password

	info, _ := json.Marshal(loginInfo)
	sendMsg(conn, "login", string(info), "")

	//开启协程处理消息
	go onMessage(conn, username)

	//发送消息
	for {
		message := lib.FmtStr(lib.ScanLine())
		if message != "" {
			sendMsg(conn, "say", message, clientToken)
		}

	}
}

//接收消息
func onMessage(conn net.Conn, username string) {
	//定义发送的消息格式
	var inputMsg struct {
		Action   string
		Messages string
	}

	//定义发送的消息格式
	for {
		data := make([]byte, 1024)
		//读取消息
		length, _ := conn.Read(data)

		if err := json.Unmarshal(data[0:length], &inputMsg); err != nil {
			continue
		}
		if inputMsg.Action == "login" {
			clientToken = inputMsg.Messages
			fmt.Println("登录成功:" + clientToken)
			continue
		}
		if inputMsg.Action == "say" {
			//屏蔽自身发送的聊天内容
			if strings.Contains(string(inputMsg.Messages), "["+username+"]: ") == false {
				fmt.Println(string(inputMsg.Messages))
			}
			continue
		}
		if inputMsg.Action == "ping" {
			sendMsg(conn, "pong", "", clientToken)
			continue

		}

	}
}

//发送消息
func sendMsg(conn net.Conn, action string, msg string, clientToken string) error {
	message := getSendMsg(action, msg, clientToken)
	_, err := conn.Write([]byte(message))
	return err
}

//获取发送的数据
func getSendMsg(action string, msg string, clientToken string) string {
	var outputMsg struct {
		Token  string
		Action string
		Params string
	}
	outputMsg.Token = clientToken
	outputMsg.Action = action
	outputMsg.Params = msg

	message, _ := json.Marshal(outputMsg)
	return string(message)

}

//检查客户端是否在线
func checkOnline(conn net.Conn) {
	for {
		//循环ping客户端
		time.Sleep(1 * time.Second)
		err := sendMsg(conn, "ping", "", "")

		if err != nil {
			fmt.Println("与服务器的连接已经断开")
			conn.Close()
			break
		}

	}

}
