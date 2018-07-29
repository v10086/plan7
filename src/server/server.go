package server

import (
	"cache"
	"dbs"
	"encoding/json"
	"fmt"
	"github.com/pborman/uuid"
	"lib"
	"net"
	"strings"
	"time"
)

func OnStartServer(port string) {
	//创建TCP监听地址
	addr, err := net.ResolveTCPAddr("tcp", port)
	lib.CheckErr(err)

	//开始监听
	listen, err := net.ListenTCP("tcp", addr)
	lib.CheckErr(err)

	fmt.Println("Server started")

	//保存客户端连接
	conns := make(map[string]net.Conn)

	//消息通道
	msg := make(chan string, 10)

	//启动协程广播消息
	go broadcast(&conns, msg)

	//处理来自服务器的命名
	go onMessage(&conns, msg)

	//检查用户是否在线
	go checkOnline(&conns, msg)

	for {
		//接收客户端连接
		conn, err := listen.Accept()
		lib.CheckErr(err)
		//处理来自客户端信息
		go onClientMessage(conn, msg, &conns)

	}
}

//当有新的连接建立
func onClientMessage(conn net.Conn, msg chan string, conns *map[string]net.Conn) {
	//初次连接提示输入登录信息
	conn.Write([]byte("connected."))
	conn.Write([]byte("Please enter login information"))

	//定义接收的消息格式
	var inputMsg struct {
		Token  string
		Action string
		Params string
	}
	//定义接收的消息格式
	var outputMsg struct {
		Action   string
		Messages string
	}

	var messages string
	data := make([]byte, 1024)
	var loginInfo struct {
		Username string
		Password string
	}

	for {
		length, err := conn.Read(data)
		if err != nil {
			conn.Close()
			break
		}
		if length > 0 {
			data[length] = 0
		}

		//解析客户端内容
		err = json.Unmarshal(data[0:length], &inputMsg)

		if err != nil {
			sendMsg(conn, "say", "输入内容格式不正确")
			continue
		}
		outputMsg.Action = inputMsg.Action

		if inputMsg.Action != "login" && inputMsg.Action != "pong" && inputMsg.Action != "ping" {
			username, err := checkAuth(inputMsg.Token)
			if err != nil {
				sendMsg(conn, "say", "身份认证不通过")
				continue
			}
			loginInfo.Username = username

		}

		switch inputMsg.Action {
		case "login":
			if err := json.Unmarshal([]byte(inputMsg.Params), &loginInfo); err != nil {
				panic(err)
			}
			token := login(conn, loginInfo.Username, loginInfo.Password)
			if token == "ko" {
				continue
			}

			//判断是否有同名昵称并保存新的客户端连接 下线旧的客户端连接
			if _, ok := (*conns)[loginInfo.Username]; ok {
				sendMsg((*conns)[loginInfo.Username], "say", "你在新客户端登录,将被踢下线")
				(*conns)[loginInfo.Username].Close()
			} else {
				sendMsg(conn, "login", token)
				//初次登录入群 提示欢迎
				messages = "[" + loginInfo.Username + "]: 加入."
				fmt.Println(messages)
				outputMsg.Action = "say"

			}

			(*conns)[loginInfo.Username] = conn

		case "say":
			messages = "[" + loginInfo.Username + "]" + ": " + inputMsg.Params
			fmt.Println(messages)

		case "ping":
			sendMsg(conn, "ping", "pong")
			continue

		case "pong":
			continue
		}

		msg <- getSendMsg(outputMsg.Action, messages)
	}

}

func checkAuth(token string) (username string, err error) {
	redis := cache.GetCon()
	username, err = redis.Get("user:auth:token:" + token).Result()
	return
}

//用户登录
func login(conn net.Conn, username, inPassword string) string {
	var password string
	db := dbs.GetCon()
	stmt, _ := db.Prepare("select password from user where username = ? limit 1")
	defer stmt.Close()

	rows, err := stmt.Query(username)
	defer rows.Close()
	if err != nil {
		sendMsg(conn, "say", "[Server]: 服务器错误.1")
		conn.Close()
		return "ko"
	}

	for rows.Next() {
		err = rows.Scan(&password)
		if err != nil {
			sendMsg(conn, "say", "[Server]: 服务器错误.2")
			conn.Close()
			return "ko"
		}

	}

	if inPassword != password {
		sendMsg(conn, "say", "[Server]: 账号或密码错误.")
		conn.Close()
		return "ko"
	}
	token := uuid.New()
	redis := cache.GetCon()
	err = redis.Set("user:auth:token:"+token, username, time.Second*7200).Err()
	if err != nil {
		panic(err)
		sendMsg(conn, "say", "[Server]: redis set 错误.")
		conn.Close()
		return "ko"
	}
	return token

}

//群广播信息
func broadcast(conns *map[string]net.Conn, msg chan string) {
	for {
		//从通道中接收消息
		data := <-msg

		//循环客户端连接并发送消息
		for key, value := range *conns {
			_, err := value.Write([]byte(data))
			if err != nil {
				delete(*conns, key)
			}
		}
	}
}

//处理服务器的命令
func onMessage(conns *map[string]net.Conn, msg chan string) {
	for {
		message := lib.ScanLine()

		//解析命令
		cmd := strings.Split(string(message), "|")
		if len(cmd) > 1 {
			cmd[0] = lib.FmtStr(cmd[0])
			cmd[1] = lib.FmtStr(cmd[1])
			switch cmd[0] {
			case "ko":
				if _, ok := (*conns)[cmd[1]]; ok {
					//关闭对应客户端连接
					(*conns)[cmd[1]].Close()
					msg <- getSendMsg("say", "[Server]: ko ["+cmd[1]+"]")
				}
			default:
				msg <- getSendMsg("say", "[Server]: "+string(message))

			}
		} else {
			msg <- getSendMsg("say", "[Server]: "+string(message))
		}

	}
}

//检查客户端是否在线
func checkOnline(conns *map[string]net.Conn, msg chan string) {
	for {
		//循环ping客户端
		time.Sleep(1 * time.Second)
		for key, value := range *conns {
			go ping(key, value, msg)
		}
	}

}

//ping 功能
func ping(user string, conn net.Conn, msg chan string) {
	err := sendMsg(conn, "ping", "")
	if err != nil {
		message := "[Server]:" + user + " 已经下线"
		fmt.Println(message)
		message = getSendMsg("say", message)

		msg <- message
	}

}

//发送消息
func sendMsg(conn net.Conn, action string, msg string) error {
	message := getSendMsg(action, msg)
	_, err := conn.Write([]byte(message))
	return err

}

//获取发送的消息
func getSendMsg(action string, msg string) string {
	var outputMsg struct {
		Action   string
		Messages string
	}
	outputMsg.Action = action
	outputMsg.Messages = msg
	message, _ := json.Marshal(outputMsg)
	return string(message)

}
