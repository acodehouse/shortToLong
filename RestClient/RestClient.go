package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

////////////////////////////////////////////////////////////////////////////

//呼叫请求
type CallbackStruct struct {
	FromClient string `json:"fromClient"`
	To         string `json:"to"`
}

type ReqMessage struct {
	CallbackStruct `json:"callback"`
}

////////////////////////////////////////////////////////////////////
//呼叫请求回复

type RespStruct struct {
	Respcode string `json:"respcode"`
}

type RespMessage struct {
	RespStruct `json:"resp"`
}

//////////////////////////////////////////////////////////////////////////

func main() {
	var callerNumber, calledNumber string
	var callTimes int

	callerNumber = "15312345678"
	calledNumber = "18512345678"
	callTimes = 1 //callTimes

	var wg sync.WaitGroup
	// 生成token
	//tokens := make(chan int, runtime.NumCPU())
	tokens := make(chan int, callTimes)
	fmt.Println("Begin Send ")
	//记录开始时间
	start_time := time.Now()

	for i := 0; i < callTimes; i++ {
		// 获取
		tokens <- 1
		wg.Add(1)
		go func() {
			/////////////////////////////////////////////////
			//开始干活


			var vurl string = "http://127.0.0.1:8080/RestProxy"

			///////////////////////////////////////////////////////////////
			//把post对象准备好
			var sendMessage ReqMessage
			sendMessage.CallbackStruct.FromClient = callerNumber
			sendMessage.CallbackStruct.To = calledNumber

			////////////////////////////////////////////////////////////////////////
			//从对象产生二进制buffer
			b, err := json.Marshal(sendMessage)
			if err != nil {
				fmt.Println("json err:", err)
			}
			//////////////////////////////////////////////////////////////////////
			client := &http.Client{}
			//复制成新buffer发送
			reqBody := bytes.NewBuffer([]byte(b))
			reqest, err := http.NewRequest("POST", vurl, reqBody)
			//fmt.Println(reqBody)
			if err != nil {
				fmt.Println(err)
				return
			}
			reqest.Close = true
			//	fmt.Println("finish make body")

			//////////////////////////////////////////////////////////////////////////
			//补充头部
			reqest.Header.Add("Accept", "application/json")
			reqest.Header.Add("Content-Type", "application/json;charset=utf-8")

			//////////////////////////////////////////////////////////////////////////
			//发送请求

			response, err := client.Do(reqest)
			if response != nil {
				defer response.Body.Close()
			}

			if err != nil {
				fmt.Println("error for send request", err.Error())
				<-tokens
				defer wg.Done()
				return
			}

			//////////////////////////////////////////////////////////////////////
			//获得请求回应
			//var respBody []byte
			if response.StatusCode == 200 {
				ioutil.ReadAll(response.Body)
				//	bodyByte, _ := ioutil.ReadAll(response.Body)
				//fmt.Println(string(bodyByte))
				fmt.Println("get correct response:", response.StatusCode)
			} else {
				fmt.Println("not 200 error:", response.StatusCode)
				return
			}
			////////////////////////////////////////////////////////////////////

			//结束干活
			// 缴回
			<-tokens
			defer wg.Done()
		}()
	}
	wg.Wait()

	//记录结束时间
	end_time := time.Now()

	var dur_time time.Duration = end_time.Sub(start_time)

	var elapsed_nano int64 = dur_time.Nanoseconds()

	fmt.Println("End send ")
	//输出执行时间，单位为毫秒。
	fmt.Println("Run Time(毫秒)", elapsed_nano/1000000)
	fmt.Println("请按回车键退出")

	var returnstr string
	fmt.Scanln(&returnstr)
}
