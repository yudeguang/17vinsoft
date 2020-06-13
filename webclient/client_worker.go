package webclient

import (
	"fmt"
	"github.com/yudeguang/17vinsoft/agentcomm"
	"net"
	"time"
)

type clientWorker struct {
	conn net.Conn
}

func (this *clientWorker) Init() {

}
func (this *clientWorker) Start() {
	pLogger.Log("开始连接到主服务器....")
	var err = this.connectServer()
	if err != nil {
		pLogger.LogExit("TCP连接到服务器:", err)
	}
	pLogger.Log("TCP已连接,开始循环读取请求...")
	//这个地方循环接收数据开始
	for {
		header, text, err := agentcomm.ReadPackage(this.conn, 60)
		if err != nil {
			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() { //超时错误,继续
				continue
			}
			pLogger.LogExit("循环读取请求失败:", err)
		}
		if header.CmdId == agentcomm.CMD_PING_REQUEST {
			//只有这个请求才回复一下，其他的都不用处理
			err = agentcomm.WritePackage(this.conn, agentcomm.CMD_PING_REPLY, createSwitchData().ToString())
			if err != nil {
				pLogger.LogExit("回复PING请求失败:", err)
			}
			pLogger.Log("回复PING请求完成...")
		} else if header.CmdId == agentcomm.CMD_REQUEST_REVERSE_CONNECT {
			pLogger.Log("收到反向连接请求,连接号:", header.ReverseId)
			session := &clsClientSession{}
			go session.Run(header.ReverseId)
		} else {
			pLogger.Log(fmt.Sprintf("读取到未处理请求,CMD=0x%X,Data:%s", header.CmdId, text))
		}
	}
}
func (this *clientWorker) connectServer() error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", serverHost, serverTCPPort), time.Second*15)
	if err != nil {
		return fmt.Errorf("连接失败:%v", err)
	}
	err = conn.SetDeadline(time.Now().Add(time.Second * 10))
	if err != nil {
		conn.Close()
		return fmt.Errorf("设置超时失败:%v", err)
	}
	conn.(*net.TCPConn).SetKeepAlive(true)
	//发送我是主链接
	switchData := createSwitchData()
	switchData.Modules = pModuleManager.GetModuleList()
	err = agentcomm.WritePackage(conn,
		agentcomm.CMD_CONNECT_MAIN,
		switchData.ToString())
	if err != nil {
		conn.Close()
		return fmt.Errorf("发送主链接请求失败:%v", err)
	}
	//等待服务端的回馈连接成功
	head, _, err := agentcomm.ReadPackage(conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("读取主链接反馈失败:%v", err)
	}
	if head.CmdId != agentcomm.CMD_SUCCESS {
		conn.Close()
		return fmt.Errorf("主链接返回错误命令:%v", head.CmdId)
	}
	//到这里就完全握手成功了
	this.conn = conn
	return nil
}
