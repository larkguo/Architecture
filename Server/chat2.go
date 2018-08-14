package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

//========================Chat Server========================//
func serverSend2Client(client net.Conn, clientMsgQueue <-chan string) {
	for msg := range clientMsgQueue {
		fmt.Fprintln(client, msg)
	}
}

func serverHandleConn(client net.Conn, broadcastMsgQueue chan string,
	enteringQueue chan chan<- string, leavingQueue chan chan<- string) {

	//每一个client创建一个clientMsgQueue通道,通道里是需要给该client转发的数据
	clientMsgQueue := make(chan string)
	go serverSend2Client(client, clientMsgQueue)

	//client加入
	enteringQueue <- clientMsgQueue
	who := client.RemoteAddr().String()
	broadcastMsgQueue <- who + " are arrived"
	fmt.Println(who + " are arrived")

	buf := make([]byte, 4096)

	for {
		lenght, err := client.Read(buf)
		if err != nil {
			//该client退出
			leavingQueue <- clientMsgQueue
			broadcastMsgQueue <- who + " are left"
			fmt.Println(who + " are left")
			client.Close()
			break
		}

		bufStr := string(buf[0:lenght])

		//该client聊天内容转发广播
		broadcastMsgQueue <- who + ": " + bufStr
	}
}

//广播和全局数据结构统一处理协程
//Don’t communicate by sharing memory, share memory by communicating
func serverBroadcast(broadcastMsgQueue chan string,
	enteringQueue <-chan chan<- string,
	leavingQueue chan chan<- string) {

	//clients map集合里记录所有client信息,map的key对应每个client的clientQueue
	clients := make(map[chan<- string]bool)
	for {
		select {
		case msg := <-broadcastMsgQueue:
			for clientMsgQueue := range clients {
				//serverSend2Client协程一直在监听clientMsgQueue,读取其中的内容，并转发给客户端
				clientMsgQueue <- msg
			}

		case clientMsgQueue := <-enteringQueue:
			clients[clientMsgQueue] = true

		case clientMsgQueue := <-leavingQueue:
			delete(clients, clientMsgQueue)
			close(clientMsgQueue) //关闭该client的channel
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

	//创建channel队列
	broadcastMsgQueue := make(chan string, 100)
	enteringQueue := make(chan chan<- string)
	leavingQueue := make(chan chan<- string)

	//启动一个转发协程
	go serverBroadcast(broadcastMsgQueue, enteringQueue, leavingQueue)

	for {
		fmt.Println("Listening ...")
		client, err := l.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		fmt.Println("Accepting ...")

		//每一个client启动一个协程
		go serverHandleConn(client, broadcastMsgQueue, enteringQueue, leavingQueue)
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