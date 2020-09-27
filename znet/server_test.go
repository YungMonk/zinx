package znet

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/YungMonk/zinx/ziface"
)

// run in terminal:
// go test -v ./znet -run=TestServer

/*
	模拟客户端
*/
func ClientTest(i uint32) {

	fmt.Println("Client Test ... start")
	//3秒之后发起测试请求，给服务端开启服务的机会
	time.Sleep(3 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("client start err, exit!")
		return
	}

	for {
		dp := NewDataPack()
		msg, _ := dp.Pack(NewMessage(i, []byte("client test message")))
		_, err := conn.Write(msg)
		if err != nil {
			fmt.Println("client write err: ", err)
			return
		}

		//先读出流中的head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err = io.ReadFull(conn, headData)
		if err != nil {
			fmt.Println("client read head err: ", err)
			return
		}

		// 将headData字节流 拆包到msg中
		msgHead, err := dp.UnPack(headData)
		if err != nil {
			fmt.Println("client unpack head err: ", err)
			return
		}

		if msgHead.GetDataLen() > 0 {
			//msg 是有data数据的，需要再次读取data数据
			msg := msgHead.(*Message)
			msg.Data = make([]byte, msg.GetDataLen())

			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("client unpack data err")
				return
			}

			fmt.Printf("==> Client receive Msg: Id = %d, len = %d , data = %s\n", msg.ID, msg.DataLen, msg.Data)
		}

		time.Sleep(time.Second)
	}
}

/*
	模拟服务器端
*/

//ping test 自定义路由
type PingRouter struct {
	BaseRouter
}

// PreHandle 处理 Connection 业务之前的钩子方法 Hook
func (p *PingRouter) PreHandle(request ziface.IRequest) {
	fmt.Println("Call Router PreHandle")
	err := request.GetConnection().SendMsg(1, []byte("before ping ....\n"))
	if err != nil {
		fmt.Println("preHandle SendMsg err: ", err)
	}
}

// Handle 处理 Connection 主业务的钩子方法 Hook
func (p *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendMsg(1, []byte("ping...ping...ping\n"))
	if err != nil {
		fmt.Println("Handle SendMsg err: ", err)
	}
}

// PostHandle 处理 Connection 业务之后的钩子方法 Hook
func (p *PingRouter) PostHandle(request ziface.IRequest) {
	fmt.Println("Call Router PostHandle")
	err := request.GetConnection().SendMsg(1, []byte("After ping .....\n"))
	if err != nil {
		fmt.Println("Post SendMsg err: ", err)
	}
}

type HelloRouter struct {
	BaseRouter
}

// Handle 处理 Connection 主业务的钩子方法 Hook
func (hr *HelloRouter) Handle(request ziface.IRequest) {
	fmt.Println("call helloRouter Handle")
	fmt.Printf("receive from client msgId=%d, data=%s\n", request.GetMsgID(), string(request.GetData()))

	err := request.GetConnection().SendMsg(2, []byte("hello zix hello Router"))
	if err != nil {
		fmt.Println(err)
	}
}

// DoConnectionBegin 当前客户端创建连接之后执行的 Hook 函数
func DoConnectionBegin(conn ziface.IConnection) {
	fmt.Println("DoConnectionBegin is Called ... ")
	err := conn.SendMsg(2, []byte("DoConnection BEGIN..."))
	if err != nil {
		fmt.Println(err)
	}
}

// DoConnectionLost 当前客户端断开连接之前执行的 Hook 函数
func DoConnectionLost(conn ziface.IConnection) {
	fmt.Println("DoConnectionLost is Called ... ")
}

func TestServer(t *testing.T) {
	//创建一个server句柄
	s := NewServer("Zinx FrameWork testing case")

	//注册链接hook回调函数
	s.SetOnConnStart(DoConnectionBegin)
	s.SetOnConnStop(DoConnectionLost)

	// 多路由
	s.AddRouter(1, &PingRouter{})
	s.AddRouter(2, &HelloRouter{})

	//	客户端测试
	go ClientTest(1)
	go ClientTest(2)

	//2 开启服务
	go s.Serve()

	select {
	case <-time.After(time.Second * 10):
		return
	}
}
