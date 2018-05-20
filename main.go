package main

import (
        "bufio"
        "fmt"
        "os"
        "strings"
)

func main() {
        if len(os.Args) != 3 {
                fmt.Println("Please input:\"chat [server|client] [:port|IP address:port]\"")
                os.Exit(-1)
        }

        if os.Args[1] == "server" {
                onStartServer(os.Args[2])
        } else if os.Args[1] == "client" {
                onStartClient(os.Args[2])
        } else {
                fmt.Println("Wrong param")
                os.Exit(-1)
        }
        fmt.Println(os.Args[1])
        fmt.Println("Hello World!")
}

//ScanLine ��ȡ����
func ScanLine() string {
        inputReader := bufio.NewReader(os.Stdin)
        input, _ := inputReader.ReadString('\n')
        return strings.Replace(input, "\n", "", -1)
}

func checkErr(err error) {
        if err != nil {
                fmt.Println(err)
                os.Exit(-1)
        }
}