package webservices

import (
	"fmt"
	"github.com/yudeguang/17vinsoft/agentcomm"
	"net"
	"sync"
)

var pConnectCache = &clsConnectCache{
	mapCache: make(map[string]*connectMain),
}

type clsConnectCache struct {
	mapCache       map[string]*connectMain //缓存的mainconnect链接
	mapLocker      sync.RWMutex            //用于mapCache的锁
	pingThreadOnce sync.Once               //执行一个循环ping的线程
}

//添加一个主链接
func (this *clsConnectCache) Add(conn net.Conn, switchData *agentcomm.TagSwitchData) {
	/*//启动心跳检查现场
	this.pingThreadOnce.Do(func(){
		go this.loopSendPingRequestThread()
	})*/
	onlyId := switchData.OnlyId
	mainConn := &connectMain{}
	mainConn.Init(conn, onlyId, switchData.RemoteAddr)
	//锁定,加入到map,并启动一个线程循环接收
	this.mapLocker.Lock()
	defer this.mapLocker.Unlock()
	if old, ok := this.mapCache[onlyId]; ok {
		delete(this.mapCache, onlyId)
		if old != nil {
			pLogger.Log(fmt.Sprintf("[%s]删除原有主连接:%s", old.remoteAddr, old.onlyid))
			old.Disconnect()
		}
	}
	//更新数据库
	pDBHelper.UpdateAgentFromMainConnect(switchData, true)
	//启动PING独立扫描线程
	go mainConn.startPingThread()
	pLogger.Log(fmt.Sprintf("[%s]建立新的主连接:%s", switchData.RemoteAddr, switchData.OnlyId))
	this.mapCache[onlyId] = mainConn
	//go loopRecvMainConnectThread(conn,switchData)
}

//删除一个
func (this *clsConnectCache) Delete(onlyid string) {
	this.mapLocker.Lock()
	defer this.mapLocker.Unlock()
	old, ok := this.mapCache[onlyid]
	if ok {
		delete(this.mapCache, onlyid)
		if old != nil {
			pLogger.Log(fmt.Sprintf("[%s]删除主连接:%s", old.remoteAddr, old.onlyid))
			old.Disconnect()
		}
	}
}

//获得一个主连接
func (this *clsConnectCache) Get(onlyid string) *connectMain {
	this.mapLocker.Lock()
	defer this.mapLocker.Unlock()
	return this.mapCache[onlyid]
}
