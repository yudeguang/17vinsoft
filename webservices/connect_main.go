package webservices

import (
	"fmt"
	"github.com/yudeguang/17vinsoft/agentcomm"
	"net"
	"time"
)

var pingIntervalSecond = 20 //间隔发送ping请求的时间

type connectMain struct {
	onlyid     string
	conn       net.Conn
	remoteAddr string
	pingTimer  *time.Timer //间隔ping的timer
}

func (this *connectMain) Init(conn net.Conn, onlyId string, remoteAddr string) {
	this.conn = conn
	this.onlyid = onlyId
	this.remoteAddr = remoteAddr
}
func (this *connectMain) Disconnect() {
	if this.conn != nil {
		this.conn.Close()
		this.conn = nil
	}
}

//启动循环心跳探测
func (this *connectMain) startPingThread() {
	this.pingTimer = time.AfterFunc(time.Second*time.Duration(pingIntervalSecond), this.doPingRequest)
	for {
		head, text, err := agentcomm.ReadPackage(this.conn)
		if err != nil {
			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() { //超时错误,继续
				pLogger.Log(fmt.Sprintf("[%s]超时错误继续", this.remoteAddr))
				continue
			}
			break
		}
		if head.CmdId == agentcomm.CMD_PING_REPLY { //回复的连接存活
			pLogger.Log(fmt.Sprintf("[%s]收到PING回复...", this.remoteAddr))
			switchData, err := agentcomm.SwitchDataFromJson(text)
			if err == nil && switchData.OnlyId != "" && switchData.SvrFlag == theServerFlag {
				if switchData.RemoteAddr == "" {
					switchData.RemoteAddr = this.conn.RemoteAddr().String()
				}
				if switchData.RemoteAddr == "" {
					switchData.HostAddr = switchData.RemoteAddr
				}
				pDBHelper.UpdateAgentFromMainConnect(switchData, false)
			}
		}
	}
	this.pingTimer.Stop()
	this.pingTimer = nil
	this.conn.Close()
	pConnectCache.Delete(this.onlyid)
}

//发送ping数据
func (this *connectMain) doPingRequest() {
	pLogger.Log(fmt.Sprintf("[%s]发送PING请求...", this.remoteAddr))
	defer func() {
		this.pingTimer.Reset(time.Second * time.Duration(pingIntervalSecond))
	}()
	var err = agentcomm.WritePackage(this.conn, agentcomm.CMD_PING_REQUEST, "")
	if err != nil {
		pLogger.Log(fmt.Sprintf("[%s]发送PING请求失败:%v", this.remoteAddr, err))
	}
}

//发送请求连接的命令
func (this *connectMain) SendConnectRequest(sequence int64) error {
	var err = agentcomm.WritePackage(this.conn, agentcomm.CMD_REQUEST_REVERSE_CONNECT, "", sequence)
	return err
}
