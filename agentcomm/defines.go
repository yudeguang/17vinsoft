package agentcomm
import "unsafe"

//定义数据通信协议头
const(
	//MAX_ONLYID_SIZE = 20;
	//MAX_SVRFLAG_SIZE = 20;
	HEAD_FLAG = 0x34567890
)
type AgentHeader struct{
	HeadSize 	int32 //头长度
	HeadFlag 	int32 //头标记
	CmdId 	 	int32 //命令号
	DataLen  	int32
	ReverseId  	int64 //反向连接ID
	Resv 	 	int64 //保留字段
}
//头长度定义
var HEAD_SIZE = int32(unsafe.Sizeof(AgentHeader{}));
//命令定义
var CMD_SUCCESS = int32(0); //成功
var CMD_ERROR = int32(-1); //错误

var CMD_CONNECT_MAIN = int32(0x1); //主链接请求链接
var CMD_CONNECT_REPLY = int32(0x2);//回复链接
var CMD_PING_REQUEST = int32(0x3); //探测请求
var CMD_PING_REPLY = int32(0x4); //探测回复

var CMD_REQUEST_REVERSE_CONNECT = int32(5) //请求反向连接
var CMD_REPLY_REVERSE_CONNECT = int32(6)  //回复反向连接

var CMD_REQUEST_EXECUTE_TASK = int32(7) //请求执行任务