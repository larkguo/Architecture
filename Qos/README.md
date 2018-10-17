
# QoS服务器端流量控制
   
   
## 1. Qos在Linux位置

### 
    Linux中的QoS分为入口(Ingress)部分和出口(Egress)部分，大多数队列(qdisc)都是用于输出流量的带宽控制，
例如HTB队列等，而输入流量只有一个队列，即Ingress队列。HTB队列的可以设置复杂的队列规则，从而灵活的控制输出流量的带宽，
		而ingress队列的功能很简单，不可指定复杂的队列规则.如果要对输入流量做复杂的带宽控制。
    Ingress和egress控制在linux网络中的位置参见 Packet flow in Netfilter and General Networking.
    TC流量控制在Linux网络中的位置如下：
![image](https://github.com/larkguo/Architecture/blob/master/Qos/data/Netfilter-packet-flow.svg.png)
    
    简化版：
![image](https://github.com/larkguo/Architecture/blob/master/Qos/data/tc-in-linux.png)

 
## 2. 服务器端upload带宽

###
    服务器不同于路由器，能把forward的数据流对应到egress输出方向的interface接口进行上传方向上的QOS流量控制，服务器上如果要对输入流量做复杂的带宽控制，可以通过Ingress队列把输入流量重定向到虚拟设备ifb，然后对虚拟设备ifb的输出流量配置HTB队列，就能达到对输入流量设置复杂的队列规则。
    SFQ队列通过一个hash函数将不同会话(如TCP 会话或者 UDP 流)分到不同的FIFO队列中，保证不会有某一个会话独占出口带宽，从而保证数据流的公平性。
    下面以samba上传带宽QoS控制为例，限制samba client公平共享上传带宽8-9mbit，配置如下：
![image](https://github.com/larkguo/Architecture/blob/master/Qos/data/upload-qos.png)

		# ensure that the ifb module is loaded 
		modprobe ifb
		
		# Clear old queuing disciplines (qdisc) on the interfaces
		tc qdisc del dev ens33 ingress
		tc qdisc del dev ifb0 root
		
		# ensure the interface ifb is up 
		ip link set dev ifb0 up
		
		# Create ingress on external interface
		tc qdisc add dev ens33 ingress 
		 
		# Forward all ingress traffic to the IFB device
		tc filter add dev ens33 parent ffff: protocol ip u32 match u32 0 0 action mirred egress redirect dev ifb0
		
		# create the root htb qdisc
		tc qdisc add dev ifb0 root handle 1:0 htb
		
		# create the parent class 1:1 with rate 8mbit-9mbit
		tc class add dev ifb0 parent 1:0 classid 1:1 htb rate 8mbit ceil 9mbit
		
		# create the child class sfq qdisc
		tc qdisc add dev ifb0 parent 1:1 handle 10:0 sfq 
		
		# create filters for each child class
		tc filter add dev ifb0 parent 1:0 protocol ip u32 match ip dport 445 0xffff flowid 1:1
		tc filter add dev ifb0 parent 1:0 protocol ip u32 match ip dport 119 0xffff flowid 1:1

 
## 3. 验证

###
    下图可看到两个samba client同时上传大小相同的文件，几乎平稳平分总带宽9mbit:
![image](https://github.com/larkguo/Architecture/blob/master/Qos/data/SMB-HTB-IFB-SFQ.png)

    下图不含sfq队列时两个samba client带宽分配没有上图均匀和平稳:
![image](https://github.com/larkguo/Architecture/blob/master/Qos/data/SMB-HTB-IFB.png)

    不含QoS控制时samba client上传文件时带宽波动更大:
![image](https://github.com/larkguo/Architecture/blob/master/Qos/data/SMB-no-QoS.png)
