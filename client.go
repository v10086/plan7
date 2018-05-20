package main

import (
        "fmt"
        "net"
        "strings"
        "encoding/json" 
        "time"
)

var clientToken string

func onStartClient(port string) {
        //创建TCP监听地址
        addr, err := net.ResolveTCPAddr("tcp", port)
        checkErr(err)

        //建立连接
        conn, err := net.DialTCP("tcp", nil, addr)
        checkErr(err)

        //首次连接时读取服务器端的提示信息
        data := make([]byte, 1024)
        conn.Read(data)
        fmt.Println(string(data))
		//ping
		go clientCheckOnline(conn)

        //输入账户名
        fmt.Print("请输入账户名:")
        username := ScanLine()
		
		//输入密码
		fmt.Print("请输入密码:")
		password := ScanLine()
		
		//登录信息
		var loginInfo struct {
			Username string
			Password string
		}

		username = replacString(username) 
		password  = replacString(password) 
	
		loginInfo.Username = username
		loginInfo.Password = password

		info,_ := json.Marshal(loginInfo)
		sendMsgClient(conn,"login",string(info),"")

        //开启协程处理消息
        go myonMessage(conn, username)

        //发送消息
        for {
				message := replacString(ScanLine()) 
				if(message != ""){
						sendMsgClient(conn,"say",message,clientToken)
				}
                
        }
}
func replacString(str string) string {
				// 去除空格  
				str = strings.Replace(str, " ", "", -1)  
				// 去除换行符  
				str  = strings.Replace(str, "\n", "", -1) 
				// 去除换行符  
				str  = strings.Replace(str, "\r", "", -1) 
				return str

}

func myonMessage(conn net.Conn, username string) {
		//定义发送的消息格式
		var inputMessage struct {
			Action string 
			Messages string
		}
		
		//定义发送的消息格式

        for {
                data := make([]byte, 1024)
                //读取消息
                length, _ := conn.Read(data)

				if err := json.Unmarshal(data[0:length], &inputMessage); err != nil { 
					continue 
				}
				if(inputMessage.Action == "login"){
					clientToken =  inputMessage.Messages
					fmt.Println("登录成功:"+clientToken)
					continue
				}
				if(inputMessage.Action == "say"){
						//屏蔽自身发送的聊天内容
						if strings.Contains(string(inputMessage.Messages), "["+username+"]: ") == false {
								fmt.Println(string(inputMessage.Messages))
						}
						continue
				}
				if(inputMessage.Action == "ping"){
						sendMsgClient(conn,"pong","",clientToken)
						continue
				
				}

        }
}


//发送消息
func sendMsgClient(conn net.Conn,action string,msg string,clientToken string) error {
		message := getSendMsgClient(action,msg,clientToken)
		_, err := conn.Write([]byte(message))
		return err 
 }

func getSendMsgClient(action string,msg string,clientToken string) string{
		var outputMessage struct {
			Token string
			Action string 
			Params string
		}
		outputMessage.Token = clientToken
		outputMessage.Action = action
		outputMessage.Params = msg

		message ,_:= json.Marshal(outputMessage)
		return string(message )

}


//检查客户端是否在线
func clientCheckOnline(conn net.Conn) {
    for {
        //循环ping客户端 
        time.Sleep(1 * time.Second)
		err := sendMsgClient(conn,"ping","","")
		
		if(err != nil){
			fmt.Println("与服务器的连接已经断开")
			conn.Close()
			break
		}

    }

}