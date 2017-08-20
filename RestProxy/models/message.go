package models

import (
	"bufio"
	"bytes"
	//"encoding/binary"
	//"fmt"
	//"sync/atomic"
	"sync"
)

type MessageHead struct {
	Sequence        int32
	BodyTotalLength int32
}

type MessageBody struct {
	CalledNumber []byte //32
	CallerNumber []byte //32
}

type Message struct {
	MessageHead
	MessageBody
}

var seqNo int32 = 0
var myLocker = &sync.Mutex{}

func fit2Length(str string, fixLen int) []byte {
	bytesBuffer := bytes.NewBufferString(str)
	spareLen := fixLen - len(str)
	for i := 0; i < spareLen; i++ {
		bytesBuffer.WriteByte(0)
	}
	return bytesBuffer.Bytes()

}

func getByte(value int32) []byte {

	b_buf := bytes.NewBuffer([]byte{})

	b_buf.WriteByte((byte)((value >> 24) & 0xFF))
	b_buf.WriteByte((byte)((value >> 16) & 0xFF))
	b_buf.WriteByte((byte)((value >> 8) & 0xFF))
	b_buf.WriteByte((byte)(value & 0xFF))

	return b_buf.Bytes()
}

func Pack(CalledNumber string,
	CallerNumber string) (retByte []byte, retSeq int32) {
	////////////////////////////////////////////////////////
	myLocker.Lock()
	var p = &Message{}
	//atomic.AddInt32(&seqNo, 1)
	//////////////////////////////////////////////////////
	//序号递增
	seqNo = seqNo + 1
	p.MessageHead.Sequence = seqNo
	myLocker.Unlock()
	//设置消息头
	p.MessageHead.BodyTotalLength = 32 * 2
	p.MessageBody.CalledNumber = fit2Length(CalledNumber, 32)
	p.MessageBody.CallerNumber = fit2Length(CallerNumber, 32)

	/////////////////////////////////////////////////////////////////////////////

	b_buf := bytes.NewBuffer([]byte{})

	bw := bufio.NewWriter(b_buf)
	bw.Write(getByte(p.MessageHead.Sequence))
	bw.Write(getByte(p.MessageHead.BodyTotalLength))
	//////////////////////////////////////////////

	bw.Write(p.MessageBody.CalledNumber)
	bw.Write(p.MessageBody.CallerNumber)

	//////////////////////////////////////////////////
	bw.Flush()
	//fmt.Println("head_body len ", len(b_buf.Bytes()))
	////////////////////////////////////////////////////////////
	//返回byte

	retByte = b_buf.Bytes()
	retSeq = p.MessageHead.Sequence
	p = nil
	return
}
