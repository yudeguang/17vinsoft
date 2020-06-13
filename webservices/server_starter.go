package webservices

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/yudeguang/17vinsoft/agentcomm"
	"github.com/yudeguang/17vinsoft/common"
	"net"
	"path/filepath"
	"time"
)

func StartService() {
	pLogger.Log("启动服务.......")
	pProxyCache.Init()
	go runBeegoServer()
	time.Sleep(500 * time.Millisecond)
	runTCPServer()
}

//启动Beego服务
func runBeegoServer() {
	pLogger.Log("启动WEB服务,端口:", common.ServerHttpPort)
	beego.BConfig.Listen.HTTPPort = common.ServerHttpPort
	beego.BConfig.AppName = "17vinsoft"
	beego.BConfig.RunMode = "dev"
	beego.BConfig.CopyRequestBody = true
	registBeegoFuncMap()
	var viewPath = "./views"
	beego.SetStaticPath("/", viewPath)
	beego.SetStaticPath("/js/", filepath.Join(viewPath, "js"))
	beego.SetStaticPath("/css/", filepath.Join(viewPath, "css"))
	//注册beego路由
	beego.AutoRouter(&AgentController{})
	beego.Router("/request/*", &ProxyController{}, "*:Proxy")
	//注册beego函数

	beego.Run()
	pLogger.LogExit("WEB服务运行结束...")
}

//启动TCP服务
func runTCPServer() {
	pLogger.Log("启动TCP服务,端口:", common.ServerTCPPort)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", common.ServerTCPPort))
	if err != nil {
		pLogger.LogExit("启动TCP服务Listen失败:", err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(50 * time.Microsecond)
				continue
			}
			pLogger.LogExit("TCP服务Accept失败:", err)
		}
		go processTCPConnect(conn)
	}
}

//处理tcp链接
func processTCPConnect(conn net.Conn) {
	var hostPort = conn.RemoteAddr().String()
	var logPrint = func(info string) {
		pLogger.Log("[" + hostPort + "]" + info)
	}
	logPrint("收到TCP连接...")
	head, text, err := agentcomm.ReadPackage(conn)
	if err != nil {
		logPrint("读取数据失败:" + err.Error())
		conn.Close()
		return
	}
	logPrint(fmt.Sprintf("读取到数据,CmdID:0x%X,ReverseId:%d,Data:%s", head.CmdId, head.ReverseId, text))
	//连接的第一个请求只可能是CMD_CONNECT_MAIN或CMD_REQUEST_REVERSE_CONNECT
	if head.CmdId == agentcomm.CMD_CONNECT_MAIN {
		switchData, err := agentcomm.SwitchDataFromJson(text)
		if err != nil {
			logPrint("请求JSON转换为SwitchData错误:" + err.Error())
			conn.Close()
			return
		}
		if switchData.OnlyId == "" || switchData.SvrFlag == "" || switchData.SvrFlag != theServerFlag {
			logPrint("OnlyId或SvrFlag为空或SvrFlag不匹配,断开链接")
			conn.Close()
			return
		}
		switchData.RemoteAddr = conn.RemoteAddr().String()
		if switchData.RemoteAddr == "" {
			switchData.HostAddr = switchData.RemoteAddr
		}
		//回复成功,并加入到缓存连接
		err = agentcomm.WritePackage(conn, agentcomm.CMD_SUCCESS, "")
		if err != nil {
			logPrint("回复主连数据错误:" + err.Error())
			conn.Close()
			return
		}
		pConnectCache.Add(conn, switchData)
		return
	} else if head.CmdId == agentcomm.CMD_REPLY_REVERSE_CONNECT {
		//反向连接来了
		ok := pProxyCache.AddReverseReponse(head.ReverseId, conn)
		if !ok {
			conn.Close()
		}
		return
	}
	logPrint(fmt.Sprintf("连接/首包数据命令类型:0x%.8X不支持,断开链接", head.CmdId))
	conn.Close()
	return
}
