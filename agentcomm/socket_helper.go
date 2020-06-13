package agentcomm
import(
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"errors"
	"io"
	"time"
)

//实现辅助类
var ErrConnClosed = errors.New("tcp connect is closed");
var defReadTimeout = int(60) //读取默认超时时间
var defWriteTimeout = int(60) //发送默认超时时间
//接收一个协议包
func ReadPackage(conn net.Conn,timeoutseconds ...int) (header *AgentHeader,replytext string, err error){
	if(conn == nil){
		err = ErrConnClosed;
		return;
	}
	//先接收一个包头长度
	var timeout = defReadTimeout;
	if len(timeoutseconds)>0 && timeoutseconds[0]>0{
		timeout = timeoutseconds[0];
	}
	err = conn.SetReadDeadline(time.Now().Add(time.Second*time.Duration(timeout)))
	if err != nil{
		return;
	}
	var readSize int64= 0;
	var buffer = bytes.NewBuffer(nil)
	readSize,err = io.CopyN(buffer,conn,int64(HEAD_SIZE));
	if err != nil{
		buffer.Reset()
		return;
	}
	if readSize != int64(HEAD_SIZE){
		err = fmt.Errorf("读取协议头长度太短:%v!=%v",readSize,HEAD_SIZE)
		return;
	}
	header = &AgentHeader{};
	err = binary.Read(buffer,binary.LittleEndian,header);
	if err != nil{
		buffer.Reset()
		err = fmt.Errorf("协议数据头转换失败:%v",err)
		return;
	}
	if header.HeadSize != HEAD_SIZE || header.HeadFlag != HEAD_FLAG{
		buffer.Reset()
		err = fmt.Errorf("协议头无效,HeadSize:%d HeadFlag:%d",header.HeadSize,header.HeadFlag);
		return;
	}
	//开始根据长度读取数据
	if header.DataLen>10*1024*1024{
		buffer.Reset()
		err = fmt.Errorf("协议头数据长度太长DataLen:%d",header.DataLen);
		return;
	}
	buffer.Reset()
	var maxSize = 2048
	for {
		//这个地方已经开始有数据了，那么最多读取20秒就要超时
		err = conn.SetReadDeadline(time.Now().Add(time.Second*20))
		//根据剩余的长度，按4096一块读取
		var remaindSize = int(header.DataLen)-buffer.Len()
		if remaindSize<=0{ //已经读取完成
			break;
		}
		if remaindSize>maxSize{
			_,err = io.CopyN(buffer,conn,int64(maxSize));
		}else{
			_,err = io.CopyN(buffer,conn,int64(remaindSize));
		}
		if err != nil{
			err = fmt.Errorf("读取数据错误:%v",err)
			return;
		}
	}
	replytext = buffer.String()
	buffer.Reset();
	return;
}
//返回一个协议包
func WritePackage(conn net.Conn,cmdid int32,text string,args ...interface{}) error{
	if(conn == nil){
		return ErrConnClosed;
	}
	var err = conn.SetWriteDeadline(time.Now().Add(time.Second*time.Duration(defWriteTimeout)));
	if err != nil{
		return err;
	}
	var replyId int64 = 0;
	if len(args)>0{
		n,ok := args[0].(int64);
		if ok{
			replyId = n;
		}
	}
	buffer := newPackage(cmdid,text,replyId)
	_,err = io.CopyN(conn,buffer,int64(buffer.Len()))
	buffer.Reset()
	return err;
}
//创建一个新的包,内部调用
func newPackage(cmdid int32,text string,replyId int64) *bytes.Buffer{
	buf := bytes.NewBuffer(nil)
	data := []byte(text)
	header := AgentHeader{}
	header.HeadSize = HEAD_SIZE;
	header.HeadFlag = HEAD_FLAG;
	header.CmdId = cmdid
	header.DataLen = int32(len(data));
	header.ReverseId = replyId;
	binary.Write(buf,binary.LittleEndian,header);
	if(len(data)>0){
		buf.Write(data)
	}
	return buf;
}