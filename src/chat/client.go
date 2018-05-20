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
        //����TCP������ַ
        addr, err := net.ResolveTCPAddr("tcp", port)
        checkErr(err)

        //��������
        conn, err := net.DialTCP("tcp", nil, addr)
        checkErr(err)

        //�״�����ʱ��ȡ�������˵���ʾ��Ϣ
        data := make([]byte, 1024)
        conn.Read(data)
        fmt.Println(string(data))
		//ping
		go clientCheckOnline(conn)

        //�����˻���
        fmt.Print("�������˻���:")
        username := ScanLine()
		
		//��������
		fmt.Print("����������:")
		password := ScanLine()
		
		//��¼��Ϣ
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

        //����Э�̴�����Ϣ
        go myonMessage(conn, username)

        //������Ϣ
        for {
				message := replacString(ScanLine()) 
				if(message != ""){
						sendMsgClient(conn,"say",message,clientToken)
				}
                
        }
}
func replacString(str string) string {
				// ȥ���ո�  
				str = strings.Replace(str, " ", "", -1)  
				// ȥ�����з�  
				str  = strings.Replace(str, "\n", "", -1) 
				// ȥ�����з�  
				str  = strings.Replace(str, "\r", "", -1) 
				return str

}

func myonMessage(conn net.Conn, username string) {
		//���巢�͵���Ϣ��ʽ
		var inputMessage struct {
			Action string 
			Messages string
		}
		
		//���巢�͵���Ϣ��ʽ

        for {
                data := make([]byte, 1024)
                //��ȡ��Ϣ
                length, _ := conn.Read(data)

				if err := json.Unmarshal(data[0:length], &inputMessage); err != nil { 
					continue 
				}
				if(inputMessage.Action == "login"){
					clientToken =  inputMessage.Messages
					fmt.Println("��¼�ɹ�:"+clientToken)
					continue
				}
				if(inputMessage.Action == "say"){
						//���������͵���������
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


//������Ϣ
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


//���ͻ����Ƿ�����
func clientCheckOnline(conn net.Conn) {
    for {
        //ѭ��ping�ͻ��� 
        time.Sleep(1 * time.Second)
		err := sendMsgClient(conn,"ping","","")
		
		if(err != nil){
			fmt.Println("��������������Ѿ��Ͽ�")
			conn.Close()
			break
		}

    }

}