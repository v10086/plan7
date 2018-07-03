package main

import (
        "fmt"
        "os"
		"client"
		"server"
)

func main() {
        if len(os.Args) != 3 {
                fmt.Println("Please input:\"chat [server|client] [:port|IP address:port]\"")
                os.Exit(-1)
        }

        if os.Args[1] == "server" {
                server.OnStartServer(os.Args[2])
        } else if os.Args[1] == "client" {
                client.OnStartClient(os.Args[2])
        } else {
                fmt.Println("Wrong param")
                os.Exit(-1)
        }
        fmt.Println(os.Args[1])
        fmt.Println("Hello World!")
}

