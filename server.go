
package main

import (
        "fmt"
        "net"
        "strings"
        "time"
		"encoding/json" 
		"github.com/zhongbo10086/database"
		"github.com/zhongbo10086/cache"
		"github.com/pborman/uuid"
)

func onStartServer(port string) {
        //����TCP������ַ
        addr, err := net.ResolveTCPAddr("tcp", port)
        checkErr(err)

        //��ʼ����
        listen, err := net.ListenTCP("tcp", addr)
        checkErr(err)

        fmt.Println("Server started")

        //����ͻ�������
        conns := make(map[string]net.Conn)

        //��Ϣͨ��
        msg := make(chan string, 10)

        //����Э�̹㲥��Ϣ
        go broadcast(&conns, msg)

        //�������Է�����������
        go onServerMessage(&conns, msg)

        //����û��Ƿ�����
        go checkOnline(&conns, msg)

        for {
                //���տͻ�������
                conn, err := listen.Accept()
                checkErr(err)
                //�������Կͻ�����Ϣ
                go onClientMessage(conn, msg, &conns)


        }
}

//�����µ����ӽ���
func onClientMessage(conn net.Conn, msg chan string, conns *map[string]net.Conn) {
        //����������ʾ�����¼��Ϣ
        conn.Write([]byte("connected."))
		conn.Write([]byte("Please enter login information"))
		
		
		//������յ���Ϣ��ʽ
		var inputMessage struct {
			Token string
			Action string 
			Params string
		}
		//������յ���Ϣ��ʽ
		var outputMessage struct {
			Action string 
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

                //�����ͻ�������
                err = json.Unmarshal(data[0:length],&inputMessage)
				
				if(err != nil){
					sendMsg(conn,"say","�������ݸ�ʽ����ȷ")
					continue
				}
				outputMessage.Action = inputMessage.Action
			
				if(inputMessage.Action !="login"&& inputMessage.Action != "pong" &&  inputMessage.Action != "ping" ){
					username,err := checkAuth(inputMessage.Token)
					if(err != nil){
						sendMsg(conn,"say","�����֤��ͨ��")
						continue
					}
					loginInfo.Username = username
				
				}

                switch inputMessage.Action {
					case "login":
						if err := json.Unmarshal([]byte(inputMessage.Params), &loginInfo); err != nil { 
							panic(err) 
						}
						token := login(conn,loginInfo.Username,loginInfo.Password)
						if(token == "ko"){
							continue
						}
						
						//�ж��Ƿ���ͬ���ǳƲ������µĿͻ������� ���߾ɵĿͻ�������
						if _, ok := (*conns)[loginInfo.Username]; ok {
								sendMsg((*conns)[loginInfo.Username],"say","�����¿ͻ��˵�¼,����������")
								(*conns)[loginInfo.Username].Close()
						} else {
							sendMsg(conn,"login",token)
							//���ε�¼��Ⱥ ��ʾ��ӭ
							messages = "[" + loginInfo.Username + "]: ����."
							fmt.Println(messages)
							outputMessage.Action = "say"

						}
						
						(*conns)[loginInfo.Username] = conn
						
						
					case "say":
							messages = "[" + loginInfo.Username + "]" + ": " + inputMessage.Params
							fmt.Println(messages)
							
					case "ping":
							sendMsg(conn,"ping","pong")
							continue
							
					case "pong":
							continue
					}
					
				msg <- getSendMsg(outputMessage.Action,messages)
        }

}

func checkAuth(token string)(username string ,err error) {
		redis := cache.GetCon()
		username, err = redis.Get("user:auth:token:"+token).Result()
		return
}
//�û���¼
func login(conn net.Conn,username,inPassword string) string {
		var password string
		db := database.GetCon()
		stmt, _ := db.Prepare("select password from user where username = ? limit 1")
		defer stmt.Close()
	
		rows, err := stmt.Query(username)
		defer rows.Close()
		if err != nil {
			sendMsg(conn,"say","[Server]: ����������.1")
			conn.Close()
			return "ko"
		}

		for rows.Next() {
			err = rows.Scan(&password)
			if err != nil {
				sendMsg(conn,"say","[Server]: ����������.2")
				conn.Close()
				return "ko"
			}

		}
		
		if(inPassword  != password){
				sendMsg(conn,"say","[Server]: �˺Ż��������.")
				conn.Close()
				return "ko"
		}
		token := uuid.New()
		redis := cache.GetCon()
		err = redis.Set("user:auth:token:"+token,username,time.Second*7200).Err()
	    if err != nil {
			panic(err)
			sendMsg(conn,"say","[Server]: redis set ����.")
			conn.Close()
			return "ko"
		}
		return token

}

//Ⱥ�㲥��Ϣ
func broadcast(conns *map[string]net.Conn, msg chan string) {
        for {
                //��ͨ���н�����Ϣ
                data := <-msg

                //ѭ���ͻ������Ӳ�������Ϣ
                for key, value := range *conns {
                        _, err := value.Write([]byte(data))
                        if err != nil {
                                delete(*conns, key)
                        }
                }
        }
}


//���������������
func onServerMessage(conns *map[string]net.Conn, msg chan string) {
        for {
                message := ScanLine()

                //��������
                cmd := strings.Split(string(message), "|")
                if len(cmd) > 1 {
						cmd[0] = replacString(cmd[0])
						cmd[1] = replacString(cmd[1])
                        switch cmd[0] {
							case "ko":
									if _, ok := (*conns)[cmd[1]]; ok {
										//�رն�Ӧ�ͻ�������
										(*conns)[cmd[1]].Close()
										msg <- getSendMsg("say","[Server]: ko [" + cmd[1] + "]")
									}
							default:
								msg <- getSendMsg("say","[Server]: " + string(message))

                        }
                } else {
						msg <- getSendMsg("say","[Server]: " + string(message))
                }

        }
}


//���ͻ����Ƿ�����
func checkOnline(conns *map[string]net.Conn, msg chan string) {
    for {
        //ѭ��ping�ͻ��� 
        time.Sleep(1 * time.Second)
        for key, value := range *conns {
                go ping(key,value,msg)
        }
    }

}
//ping ����
func ping(user string,conn net.Conn,msg chan string) {
        err := sendMsg(conn,"ping","")
        if err != nil {
			message := "[Server]:"+user+" �Ѿ�����"
			fmt.Println(message)
			message = getSendMsg("say",message)
			
            msg <- message
        }

}
//������Ϣ
func sendMsg(conn net.Conn,action string,msg string) error {
		message:= getSendMsg(action,msg)
		_, err := conn.Write([]byte(message))
		return err 
			
		
 }
 func getSendMsg(action string,msg string) string {
 		var outputMessage struct {
			Action string 
			Messages string
		}
		outputMessage.Action = action
		outputMessage.Messages = msg
		message ,_:= json.Marshal(outputMessage)
		return string(message )
 
 }