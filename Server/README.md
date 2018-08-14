
# 服务器协程/线程模型

## 1. worker-pool模型

### 
    go语言原生支持goroutine协程,但随着并发量增加,频繁创建和销毁协程对系统压力也会变大,预先创建worker协程池/线程池是服务器模型的通用做法.
    go语言自带线程安全的channel,类似于其他语言的queue或linux的pipe管道,用于dispatch调度协程和worker工作协程之间通信.
    该模型适用于worker协程之间相互独立,没有影响,各自处理过程可阻塞等待,不用异步处理,逻辑和层次简单.
    以网络请求为例,关键处理过程：
      1.服务器侦听到client连接后把该连接的处理handleRequest()加到dispatcher调度器的JobQueue(channel)的job callback回调里;
      2.dispatcher从JobQueue取出job，再从workerPool(channel)里取出空闲worker，并把job填进该worker的jobChannel;
      3.该worker从自己的jobChannel里取出job回调执行.
代码见[worker-pool.go](https://github.com/larkguo/Architecture/blob/master/Server/worker-pool.go),
模型如下：
![image](https://github.com/larkguo/Architecture/blob/master/Server/worker-pool.png)   


## 2. 分层模型

### 
    该模型适用于处理复杂,需要不同层次分别处理，每个层次处理一部分内容,以一个聊天服务器为例,关键处理过程：
      1.有client加入或离开时,把client信息加入enteringQueue(channel)或leavingQueue队列,待广播协程接收和维护客户信息;
      2.每个client加入后启动单独协程serverHandleConn()处理,接收msg消息,写入broadcastMsgQueue(channel)广播队列;
      3.全局广播协程serverBroadcast(),维护全局变量clients客户信息,并从broadcastMsgQueue里接收msg消息进行广播.
  
代码见[chat1.go](https://github.com/larkguo/Architecture/blob/master/Server/chat1.go),
模型如下：
![image](https://github.com/larkguo/Architecture/blob/master/Server/chat1-thread.png)  
 
 
    上面的模型中serverBroadcast广播协程response会阻塞，影响其他client请求的即时处理,
    下面做一个改进:每个client使用serverHandleConn请求和serverSend2Client响应两个协程.关键过程如下：
      1.每个client关联一个clientMsgQueue(channel)，有client加入时把该clientMsgQueue加入到全局变量clients里;
      2.广播消息时取出所有client的clientMsgQueue,并把msg消息写入;
      3.每个client的serverSend2Client响应协程从自己的clientMsgQueue取出msg消息并发送.
    
代码见[chat2.go](https://github.com/larkguo/Architecture/blob/master/Server/chat2.go),
模型如下：
![image](https://github.com/larkguo/Architecture/blob/master/Server/chat2-thread.png)  

    上面模型进一步抽象为通用分层模型,serverHandleConn和serverSend2Client抽象为低层处理,serverBroadcast处理抽象为高层处理.
    
![image](https://github.com/larkguo/Architecture/blob/master/Server/chat2-abstract.png)  

## 3. 混合模型
### 
    前两中模型的混合使用，上面主控分层模型中，client的请求和响应协程不用临时创建，使用worker-pool模型事先创建好，你可以试试！

