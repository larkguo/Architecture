package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

//========================Chat Server========================//
func serverHandleConn(client net.Conn,
	broadcastMsgQueue chan string, leavingQueue chan string) {

	buf := make([]byte, 4096)
	who := client.RemoteAddr().String()
	for {
		lenght, err := client.Read(buf)
		if err != nil {
			//该client退出
			leavingQueue <- who
			break
		}

		bufStr := string(buf[0:lenght])
		fmt.Println("<--", who, bufStr)

		//该client聊天内容转发广播
		broadcastMsgQueue <- bufStr
	}
}

//广播和关键数据结构统一处理协程
func serverBroadcast(broadcastMsgQueue chan string, enteringQueue <-chan net.Conn, leavingQueue chan string) {

	clients := make(map[string]net.Conn)
	for {
		select {
		case msg := <-broadcastMsgQueue:
			for who, client := range clients {
				_, err := client.Write([]byte(msg)) //阻塞
				if err != nil {
					fmt.Println(err.Error())
					leavingQueue <- who
				} else {
					fmt.Println("-->", who, string(msg))
				}
			}
		case client := <-enteringQueue:
			who := client.RemoteAddr().String()
			clients[who] = client
			broadcastMsgQueue <- who + " are arrived"
		case who := <-leavingQueue:
			client := clients[who]
			delete(clients, who)
			client.Close()
			broadcastMsgQueue <- who + " are left"
		}
	}
}

func ServerStart(listenAddr string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", listenAddr)
	if err != nil {
		log.Panic(err)
	}
	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Panic(err)
	}
	defer l.Close()

	broadcastMsgQueue := make(chan string, 100)
	leavingQueue := make(chan string, 10)
	enteringQueue := make(chan net.Conn, 10)

	//启动一个转发协程, make返回类型的引用
	go serverBroadcast(broadcastMsgQueue, enteringQueue, leavingQueue)

	for {
		fmt.Println("Listening ...")
		client, err := l.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		fmt.Println("Accepting ...")

		//client加入
		enteringQueue <- client

		//每个client启动一个协程
		go serverHandleConn(client, broadcastMsgQueue, leavingQueue)
	}
	close(broadcastMsgQueue)
	close(leavingQueue)
	close(enteringQueue)
}

//========================Chat Client========================//
func clientSend2Server(conn net.Conn) {
	var input string
	for {
		fmt.Scanln(&input)
		if input == "/quit" {
			conn.Close()
			log.Fatal("ByeBye..")
		}

		lens, err := conn.Write([]byte(input))
		fmt.Println(lens)
		if err != nil {
			fmt.Println(err.Error())
			conn.Close()
			break
		}
	}
}

func ClientStart(serverAddr string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", serverAddr)
	if err != nil {
		log.Panic(err)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Panic(err)
	}

	go clientSend2Server(conn)

	buf := make([]byte, 4096)
	for {
		lenght, err := conn.Read(buf)
		if err != nil {
			conn.Close()
			log.Panic(err)
		}
		fmt.Println(string(buf[0:lenght]))
	}
}

//========================Chat Main========================//
//server: chat server [ServerIp]:[ServerPort]  eg: ./chat server :9999
//client: chat client [ServerIp]:[ServerPort]  eg: ./chat client 127.0.0.1:9999
func main() {
	if len(os.Args) != 3 {
		log.Panic("please check params!")
	}
	if os.Args[1] == "server" && len(os.Args) == 3 {
		ServerStart(os.Args[2])
	}
	if os.Args[1] == "client" && len(os.Args) == 3 {
		ClientStart(os.Args[2])
	}
}
