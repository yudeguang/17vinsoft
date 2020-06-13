package webservices

import (
	"bytes"
	"fmt"
	"github.com/yudeguang/17vinsoft/agentcomm"
	"github.com/yudeguang/17vinsoft/common"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type tagValidAgentItem struct {
	OnlyId    string
	ChsName   string
	Priority  int
	ModName   string
	Timeout   string
	StaticVar string
}

//代理控制请求类
var plGlobalProxySequence = new(int64)

type ProxyController struct {
	BaseController
	moduleName string        //当前请求的moduleName
	headPart   string        //记录到日志里的前缀部分
	output     *bytes.Buffer //收到的数据
}

func (this *ProxyController) logInfo(args ...interface{}) {
	pLogger.Log(this.headPart + fmt.Sprint(args...))
}

//代理一个请求
func (this *ProxyController) Proxy() {
	var err error = nil
	this.output = bytes.NewBuffer(nil)
	//退出时根据是否有错误返回
	defer func() {
		if err != nil {
			this.logInfo("错误返回:", err.Error())
			this.Ctx.WriteString("ERR:" + err.Error())
		} else {
			this.logInfo("正确返回,数据长度:", this.output.Len())
			this.Ctx.WriteString(this.output.String())
		}
	}()
	this.moduleName = this.Ctx.Input.Param(":splat")
	this.moduleName = strings.ToLower(this.moduleName)
	this.headPart = fmt.Sprintf("[proxy:%s:%s]", this.Ctx.Request.RemoteAddr, this.moduleName)
	if this.moduleName == "" {
		err = fmt.Errorf("参数为空")
		return
	}
	this.logInfo("收到代理请求...")
	var lstItem = []*tagValidAgentItem{}
	lstItem, err = this.getValidAgentItem(this.moduleName)
	if err != nil {
		err = fmt.Errorf("查询可用Agent失败:%v", err)
		return
	}
	if len(lstItem) == 0 {
		err = fmt.Errorf("没用可用的Agent代理:%s", this.moduleName)
		return
	}
	//这里找到了配置的可用代理，还要确定哪个代理不忙，选不忙的代理使用
	lstItem = pProxyCache.FilterValidAgent(lstItem)
	if len(lstItem) == 0 {
		err = fmt.Errorf("所有能代理:%v 的Agent都在使用中", this.moduleName)
		return
	}
	//一个一个代理发送数据过去,看看谁能回应
	var pValidAgent *tagValidAgentItem = nil
	var newConn net.Conn = nil
	for _, p := range lstItem {
		this.logInfo("尝试建立与OnlyId:", p.OnlyId, "之间的代理...")
		newConn, err = this.tryCreateRevserveConnect(p)
		if err == nil {
			pValidAgent = p
			break
		} else {
			this.logInfo("建立与OnlyId:", p.OnlyId, "的代理失败:", err)
		}
	}
	if newConn == nil {
		err = fmt.Errorf("所有可用OnlyId都尝试无法建立连接")
		return
	}
	this.logInfo("与OnlyId:", pValidAgent.OnlyId, "建立代理成功...")
	err = this.executeTask(newConn, pValidAgent)
	//最后要释放占用
	newConn.Close()
	pProxyCache.UnlockAgentModule(pValidAgent.OnlyId, pValidAgent.ModName)
}

//代理一个请求
func (this *ProxyController) getValidAgentItem(modName string) ([]*tagValidAgentItem, error) {
	var sqlText = "SELECT A.OnlyId,A.ChsName,A.Priority,B.Name,B.Timeout,B.StaticVar FROM AgentList AS A INNER JOIN AgentModule AS B ON A.OnlyId=B.OnlyId"
	sqlText += " WHERE B.IsValid=1 AND B.Name=? ORDER BY A.Priority DESC"
	rows, err := pDBHelper.GetDB().Query(sqlText, modName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lstResult = []*tagValidAgentItem{}
	for rows.Next() {
		p := &tagValidAgentItem{}
		err = rows.Scan(&p.OnlyId,
			&p.ChsName,
			&p.Priority,
			&p.ModName,
			&p.Timeout,
			&p.StaticVar)
		if err != nil {
			return nil, err
		}
		lstResult = append(lstResult, p)
	}
	return lstResult, nil
}

//发送一个代理请求,看看是否能回应
func (this *ProxyController) tryCreateRevserveConnect(p *tagValidAgentItem) (net.Conn, error) {
	var err = pProxyCache.LockAgentModule(p.OnlyId, p.ModName)
	if err != nil {
		return nil, err
	}
	connMain := pConnectCache.Get(p.OnlyId)
	if connMain == nil {
		pProxyCache.UnlockAgentModule(p.OnlyId, p.ModName)
		return nil, fmt.Errorf("没有在ConnectCache中发现OnlyId:%v", p.OnlyId)
	}
	//发送反向连接请求
	revserveId := atomic.AddInt64(plGlobalProxySequence, 1)
	//先准备一个管道放在缓存里再发送数据
	cChannel := make(chan net.Conn)
	pProxyCache.AddReverseRequest(revserveId, cChannel)
	defer func() {
		pProxyCache.DelReverseRequest(revserveId)
	}()
	//发送反向连接请求
	if err = connMain.SendConnectRequest(revserveId); err != nil {
		pProxyCache.UnlockAgentModule(p.OnlyId, p.ModName)
		return nil, err
	}
	//等待接收，并且删除请求
	var newConn net.Conn = nil
	select {
	case newConn = <-cChannel:
	case <-time.After(time.Second * 3):
	}
	if newConn == nil {
		pProxyCache.UnlockAgentModule(p.OnlyId, p.ModName)
		return nil, fmt.Errorf("等待反向连接超时")
	}
	return newConn, nil
}

//向Agent发送请求执行数据
func (this *ProxyController) executeTask(conn net.Conn, pModule *tagValidAgentItem) error {
	//准备发送的参数
	param := &agentcomm.TagTaskParam{}
	param.TaskName = this.moduleName
	param.TimeoutSecond, _ = strconv.Atoi(pModule.Timeout)
	if param.TimeoutSecond <= 0 {
		param.TimeoutSecond = 60
	}
	lines := []string{}
	lines = append(lines, "#control vars")
	lines = append(lines, fmt.Sprintf("TimeoutSecond=%d", param.TimeoutSecond))
	lines = append(lines, "#static config vars")
	for _, s := range strings.Split(pModule.StaticVar, "\n") {
		s = strings.TrimSpace(s)
		lines = append(lines, s)
	}
	lines = append(lines, "#user request vars")
	for key, lst := range this.Input() {
		var value = ""
		if len(lst) > 0 {
			value = lst[0]
		}
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}
	param.LstLine = lines
	jsonText := common.ToJsonString(param)
	err := agentcomm.WritePackage(conn, agentcomm.CMD_REQUEST_EXECUTE_TASK, jsonText)
	if err != nil {
		return fmt.Errorf("发送任务请求到Agent错误:%v", err)
	}
	//等待接收数据,超时必等待时间多个10s
	header, text, err := agentcomm.ReadPackage(conn, param.TimeoutSecond+10)
	if err != nil {
		return fmt.Errorf("从Agent接收任务结果失败:%v", err)
	}
	if header.CmdId != agentcomm.CMD_SUCCESS {
		return fmt.Errorf("Agent返回错误CmdId:0X%.X,%s", header.CmdId, text)
	}
	this.output.WriteString(text)
	return nil
}
