package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
)

//===================================pool===================================
// refer https://github.com/ivpusic/grpool

// Represents user request, function which should be executed in some worker.
type Job func() //callback()

// Gorouting instance which can accept client jobs
type worker struct {
	id 	int
	workerPool chan *worker
	jobChannel chan Job
	stop       chan struct{}
}
func (w *worker) start() {
	go func() {
		var job Job
		for {
			w.workerPool <- w // worker创建或job()回调执行完后添加到空闲队列workerPool里
			select {
			case job = <-w.jobChannel:
				fmt.Println("worker:",w.id)
				job() // callback回调, 可以block阻塞执行
			case <-w.stop: // 收到退出通知
				w.stop <- struct{}{} // 反馈给dispatcher:完成退出
				return
			}
		}
	}()
}
func newWorker(pool chan *worker,workerId int) *worker {
	return &worker{
		id:workerId,
		workerPool: pool,
		jobChannel: make(chan Job),
		stop:       make(chan struct{}),
	}
}

// Accepts jobs from clients, and waits for first free worker to deliver job
type dispatcher struct {
	workerPool chan *worker
	jobQueue   chan Job
	stop       chan struct{}
}
func (d *dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue: //来活了: 从workerPool取出空闲worker协程，把任务分给该worker
			worker := <-d.workerPool
			worker.jobChannel <- job
		case <-d.stop: // 收到退出通知
			for i := 0; i < cap(d.workerPool); i++ {
				worker := <-d.workerPool
				fmt.Println("worker stop,", cap(d.workerPool))
				worker.stop <- struct{}{} // 通知worker退出
				<-worker.stop // 等待worker退出完成的反馈
			}
			d.stop <- struct{}{} // 反馈:完成退出
			return
		}
	}
}
func newDispatcher(workerPool chan *worker, jobQueue chan Job) *dispatcher {
	d := &dispatcher{
		workerPool: workerPool,
		jobQueue:   jobQueue,
		stop:       make(chan struct{}),
	}
	for i := 1; i <= cap(d.workerPool); i++ { //启动多个worker协程，填满workerPool
		worker := newWorker(d.workerPool,id)
		worker.start()
	}
	go d.dispatch() //开启调度协程
	return d
}

type Pool struct {
	JobQueue   chan Job
	dispatcher *dispatcher
}

// Will make pool of gorouting workers.
// numWorkers - how many workers will be created for this pool
// queueLen - how many jobs can we accept until we block
//
// Returned object contains JobQueue reference, which you can use to send job to pool.
func NewPool(numWorkers int, jobQueueLen int) *Pool {
	jobQueue := make(chan Job, jobQueueLen)
	workerPool := make(chan *worker, numWorkers)
	pool := &Pool{
		JobQueue:   jobQueue,
		dispatcher: newDispatcher(workerPool, jobQueue),
	}
	return pool
}

// Will release resources used by pool
func (p *Pool) Release() {
	p.dispatcher.stop <- struct{}{} // 通知dispatcher退出
	<-p.dispatcher.stop // 等待dispatcher退出完成的反馈
}

//===================================test===================================
var (
	numWorkers  int = 2
	jobQueueLen int = 1
)

const (
	CONN_HOST = "" //localhost
	CONN_PORT = "9999"
	CONN_TYPE = "tcp"
)

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	reqLen, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	conn.Write(buf[:reqLen])
	conn.Close()
}

func main() {
	pool := NewPool(numWorkers, jobQueueLen) // number of workers, and size of job queue
	defer pool.Release()

	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		conn, err := l.Accept() // Listen for an incoming connection.
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		} else {
			fmt.Println("<< ", conn.RemoteAddr())
			pool.JobQueue <- func() { // 每个client启动一个协程，各自处理，互相没有交互
				handleRequest(conn) // job回调,可以是任意形式
			}
		}
	}
}
