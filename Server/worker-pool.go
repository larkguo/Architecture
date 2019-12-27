package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
)

//===================================pool===================================
//refer https://github.com/ivpusic/grpool

// Represents user request, function which should be executed in some worker.
type Job func() //callback()

// Gorouting instance which can accept client jobs
type worker struct {
	workerPool chan *worker
	jobChannel chan Job
	stop       chan struct{}
}

func (w *worker) start() {
	go func() {
		var job Job
		for {
			// worker创建或job()回调执行完后添加到空闲队列workerPool里
			w.workerPool <- w

			select {
			case job = <-w.jobChannel:
				job() // callback回调, 可以block阻塞执行
			case <-w.stop: // 收到退出通知
				w.stop <- struct{}{} // 反馈给dispatcher:完成退出
				return
			}
		}
	}()
}

func newWorker(pool chan *worker) *worker {
	return &worker{
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
		case job := <-d.jobQueue:

			//从workerPool取出空闲worker协程，把任务分给该worker
			worker := <-d.workerPool
			worker.jobChannel <- job

		case <-d.stop:
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

	//启动多个worker协程，填满workerPool
	for i := 0; i < cap(d.workerPool); i++ {
		worker := newWorker(d.workerPool)
		worker.start()
	}

	go d.dispatch()
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
	maxJobs     int = 8
	numWorkers  int = 10
	jobQueueLen int = 5
)

const (
	CONN_HOST = "" //localhost
	CONN_PORT = "9999"
	CONN_TYPE = "tcp"
)

func init() {
	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)
	println("using MAXPROC", numCPUs)
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {

	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	// Send a response back to person contacting us.
	conn.Write(buf[:reqLen])

	// Close the connection when you're done with it.
	conn.Close()
}

func main() {
	// number of workers, and size of job queue
	pool := NewPool(numWorkers, jobQueueLen)
	defer pool.Release()

	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()

	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		} else {
			fmt.Println("<< ", conn.RemoteAddr())

			// 每个client启动一个协程，各自处理，互相没有交互
			pool.JobQueue <- func() {

				handleRequest(conn) // job回调
			}
		}
	}
}
