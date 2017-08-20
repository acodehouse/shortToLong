package controllers

import (
	"RestProxy/models"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/utils"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

var remoteAddr string = "127.0.0.1:8081"

/////////////////////////////////////////////////////
type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	c.Data["Website"] = "beego.me"
	c.Data["Email"] = "astaxie@gmail.com"
	c.TplNames = "index.tpl"
}

///////////////////////////////////////////////////////////////////////////
//长连接呼叫请求
type Request struct {
	reqSeq  int32
	reqPkg  []byte
	rspChan chan []byte
}

////////////////////////////////////////////
//短连接呼叫请求
type CallbackStruct struct {
	FromClient string `json:"fromClient"`
	To         string `json:"to"`
}

type ReqMessage struct {
	CallbackStruct `json:"callback"`
}

//////////////////////////////////////////
//处理短连接呼叫请求的Controller

var g_reqChan chan *Request //用于和长连接通信的chan

type PackRequest struct {
	request *Request
}

var g_conn *net.TCPConn //保存TCP连接

type RestProxyController struct {
	beego.Controller
}

type CallbackRespMessage struct {
	ReturnCode string `json:"returnCode"`
	ReturnDesc string `json:"returnDesc,omitempty"`
	ReturnData string `json:"returnData,omitempty"`
}

////////////////////////////////////////////////////////////////////////////
//建立短连接请求的chan
func (c *RestProxyController) MyInit() {
	g_reqChan = make(chan *Request, 1000)

	fmt.Println("MyInit")
	go c.connHandler()

}

func getInt32(value []byte) int32 {
	var ret int32
	ret = int32(value[0])<<24 + int32(value[1])<<16 + int32(value[2])<<8 + int32(value[3])
	return ret
}

////////////////////////////////////////////////////
//长连接程序,有关于长连接发送和接收通道
func (c *RestProxyController) connHandler() {
	addr, err := net.ResolveTCPAddr("tcp4", remoteAddr)
	if err != nil {
		fmt.Println("begin net.ResolveUDPAddr fail.", err)
		beego.Info("begin net.ResolveUDPAddr fail.", err)
		return
	}
	fmt.Println("begin net.ResolveTCPAddr")
	beego.Info("begin net.ResolveTCPAddr ")
	connRet, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Println("begin net.DialTCP fail.", err)
		return
	}
	//保存连接
	g_conn = connRet

	fmt.Println("begin net.DialTCP")
	beego.Info("begin net.DialTCP")

	defer connRet.Close()

	sendChan := make(chan []byte, 1000)
	go sendHandler(sendChan)

	recvChan := make(chan []byte, 1000)
	go recvHandler(recvChan)
	//////////////////////////////////////////////

	//带锁的Map
	reqMap := utils.NewBeeMap()
	for {
		select {
		//请求来了
		case req := <-g_reqChan:
			//fmt.Println("begin my send")
			//beego.Info("NormalRequest recv ", int(req.reqSeq))

			//在map里保存
			beego.Info("Request seq is ", req.reqSeq)
			reqMap.Set(req.reqSeq, req)

			//beego.Info("begin my send,mapSize:", len(reqMap.Items()))
			sendChan <- req.reqPkg
			//fmt.Println("NormalRequest recv", int(req.reqSeq), ",", string(req.reqPkg)))
		case rsp := <-recvChan:
			//获得序号
			conv := rsp[0:4]
			seq := getInt32(conv)
			//beego.Info("try find request,map size:", len(reqMap.Items()))
			//在map里获得
			//beego.Info("reqMap.Get ", seq)
			temp := reqMap.Get(seq)
			if temp == nil {
				fmt.Println("seq not found. seq=", seq, ",rsp[8]:", rsp[8], ",rsp[9]:", rsp[9], ",rsp[10]:", rsp[10], ",rsp[11]:", rsp[11])
				beego.Info("seq not found. seq=", seq, ",rsp[8]:", rsp[8], ",rsp[9]:", rsp[9], ",rsp[10]:", rsp[10], ",rsp[11]:", rsp[11])
				continue
			}

			req := temp.(*Request)
			if req == nil {
				fmt.Println("seq not found. seq=", seq)
				beego.Info("seq not found. seq=", seq)
				continue
			}
			//通知返回响应
			req.rspChan <- rsp
			//fmt.Println("send rsp to client. rsp=", string(rsp))
			//beego.Info("send rsp to client. rsp=", string(rsp))

			//在Map里清除相关数据
			//beego.Info("reqMap.delete ", req.reqSeq)
			reqMap.Delete(req.reqSeq)
			//获得map大小
			//beego.Info("return rsp to client,map size:", len(reqMap.Items()))

			//default:
			///fmt.Println("channel is full !")
		}
	}

}

////////////////////////////////////////////////////////////

var Locker sync.Mutex

//再次建立连接接收一次
func ReadAgain(b []byte) (n int, err error) {
	Locker.Lock()
	defer Locker.Unlock()
	n, err = g_conn.Read(b) //双重检测
	if err == nil {         //已经没问题了，不需要再来
		beego.Info("ReadAgain no error.")
		return
	}
	g_conn.Close()
	g_conn = nil
	var addr *net.TCPAddr
	addr, err = net.ResolveTCPAddr("tcp4", remoteAddr)
	if err != nil {
		fmt.Println("Rebuild net.ResolveUDPAddr fail.", err)
		beego.Info("Rebuild net.ResolveUDPAddr fail.", err)
		return
	}
	g_conn, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Println("Rebuild net.DialTCP fail.", err)
		beego.Info("Rebuild net.DialTCP fail.", err)
		return
	}

	n, err = g_conn.Read(b)

	return
}

//再次建立连接发送一次
func WriteAgain(b []byte) (n int, err error) {
	Locker.Lock()
	defer Locker.Unlock()
	n, err = g_conn.Write(b) //双重检测
	if err == nil {          //已经没问题了，不需要再来
		return
	}
	g_conn.Close()
	g_conn = nil
	var addr *net.TCPAddr
	addr, err = net.ResolveTCPAddr("tcp4", remoteAddr)
	if err != nil {
		fmt.Println("Rebuild net.ResolveUDPAddr fail.", err)
		beego.Info("Rebuild net.ResolveUDPAddr fail.", err)
		return
	}

	g_conn, err = net.DialTCP("tcp", nil, addr)

	if err != nil {
		fmt.Println("Rebuild net.DialTCP fail.", err)
		beego.Info("Rebuild net.DialTCP fail.", err)
		return
	}

	n, err = g_conn.Write(b)

	return
}

//////////////////////////////////////////////////////////////
//长连接发送
func sendHandler(sendChan <-chan []byte) {
	for data := range sendChan {
		wlen, err := g_conn.Write(data)
		if err == io.EOF {
			wlen, err = WriteAgain(data)
		}

		if err != nil || wlen != len(data) {
			beego.Info("conn.Write fail.", err)
			continue
		}

	}
}

////////////////////////////////////////////////////////////////////
//长连接接收
func recvHandler(recvChan chan<- []byte) {
	for {
		buf := make([]byte, 1024)
		rlen, err := g_conn.Read(buf)
		if err == io.EOF {
			rlen, err = ReadAgain(buf)
		}
		if err != nil || rlen <= 0 {
			//fmt.Println(err.Error())
			beego.Info("conn.Read fail,need rebuild conn.", err)
			continue
		}

		recvChan <- buf[:rlen]
	}
}

//////////////////////////////////////////////////////////////////////////
//获取网址参数，填入对象，写入buffer,构造消息加队列,用Seq做Key
//等待消息，返回json
func (c *RestProxyController) Call() {
	var retMsg CallbackRespMessage
	var reqMsg ReqMessage

	fmt.Println(c.Ctx.Input.RequestBody)
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &reqMsg)
	if err != nil {
		fmt.Println("Unmarshal error")
		return
	}

	calledNumber := reqMsg.To
	callerNumber := reqMsg.FromClient

	beego.Info("calledNumber:" + calledNumber)
	beego.Info("callerNumber:" + callerNumber)

	//myMessage := new(models.Message)
	//结构转换成二进制Buffer,获得序号
	sendBuffer, seq := models.Pack(calledNumber, callerNumber)
	////////////////////////////////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////////////////////////////
	rspChan := make(chan []byte, 1)
	//fmt.Println("make buffer finish ", seq, ",", len(sendBuffer))
	//beego.Info("make buffer finish ", seq, ",", len(sendBuffer))
	/////////////////////////////////////////////////////////////////
	//触发消息
	g_reqChan <- &Request{seq, sendBuffer, rspChan}
	////////////////////////////////////////////////////
	select {
	case rsp := <-rspChan:
		//fmt.Println("recv rsp. rsp", string(rsp))
		//beego.Info("recv rsp")
		//获得状态码并转换为字符串
		retMsg.ReturnCode = strconv.Itoa(int(rsp[8]))
		c.Data["json"] = &retMsg
		c.ServeJson()
		return

	case <-time.After(60 * time.Second):
		fmt.Println("wait for rsp timeout.")
		retMsg.ReturnCode = "1"
		retMsg.ReturnDesc = "ECC 连接超时"
		c.Data["json"] = &retMsg
		c.ServeJson()
		return
	}

	///////////////////////////////////////////////////////////////////////////

	retMsg.ReturnCode = "0"
	c.Data["json"] = &retMsg
	c.ServeJson()
	return
}

////////////////////////////////////////////////////////////////////////////////
