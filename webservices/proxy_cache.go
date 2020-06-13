package webservices

import (
	"fmt"
	"net"
	"sync"
)

var pProxyCache = &clsProxyCache{}
//当前正在代理的缓存类
type clsProxyCache struct{
	mapProxy map[string]string;
	mapLocker sync.Mutex;
	//以下是反向连接的存储
	mapReverseConn map[int64]chan net.Conn
	mapReverseLocker sync.Mutex;
}
func (this* clsProxyCache) Init(){
	this.mapProxy = make(map[string]string)
	this.mapReverseConn = make(map[int64]chan net.Conn)
}

func (this* clsProxyCache) FilterValidAgent(lstIn []*tagValidAgentItem) []*tagValidAgentItem{
	return lstIn;
}

func (this* clsProxyCache) LockAgentModule(onlyId string,modName string) error{
	var key = onlyId+"_"+modName;
	this.mapLocker.Lock()
	defer this.mapLocker.Unlock();
	_,ok := this.mapProxy[key];
	if ok{ //已经被别人用了不能再用
		return fmt.Errorf("LockAgentModule:OnlyId:%v,Module:%s 正在使用中",onlyId,modName)
	}
	this.mapProxy[key] = "";
	return nil;
}
func (this* clsProxyCache) UnlockAgentModule(onlyId string,modName string){
	var key = onlyId+"_"+modName;
	this.mapLocker.Lock()
	defer this.mapLocker.Unlock();
	delete(this.mapProxy,key);
}

func (this* clsProxyCache) AddReverseRequest(seq int64,ch chan net.Conn){
	this.mapReverseLocker.Lock()
	defer this.mapReverseLocker.Unlock();
	this.mapReverseConn[seq] = ch;
}
func (this* clsProxyCache) DelReverseRequest(seq int64){
	this.mapReverseLocker.Lock()
	defer this.mapReverseLocker.Unlock();
	c,ok := this.mapReverseConn[seq];
	if ok{
		delete(this.mapReverseConn,seq);
		go func(ch chan net.Conn){
			if ch!=nil{
				select{
				case conn := <-ch:
					if(conn !=nil){
						conn.Close()
					}
				default:
					//donothing
				}
			}
		}(c);
	}

}
func (this* clsProxyCache) AddReverseReponse(seq int64,conn net.Conn) bool{
	this.mapReverseLocker.Lock()
	defer this.mapReverseLocker.Unlock();
	ch,ok := this.mapReverseConn[seq];
	if !ok{
		return false;
	}
	select{
		case ch<-conn:
			ok = true;
		default:
			ok = false;
			break;
	}
	return ok;
}