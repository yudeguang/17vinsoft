package webclient

import (
	"fmt"
	"github.com/yudeguang/17vinsoft/common"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	//"log"
	"path/filepath"
	"time"
)

//客户端的唯一id
var globalOnlyId string = ""

//服务端地址和标记
var serverAddr string = ""
var serverHost string = ""
var serverHttpPort string = "9000"
var serverTCPPort string = "9001"
var globalSvrFlag string = ""
var globalStartTime string = "" //程序的启动时间
var globalProcessId string = "" //进程ID号
//日志对象
var pLogger *common.BasicLogger = nil

//保存客户端的配置
var mapConfig map[string]string = nil

//client相关运行环境初始化
func InitWebClientEnv() {
	globalStartTime = common.GetNowTime()
	globalProcessId = fmt.Sprint(os.Getpid())
	pLogger = common.NewLogger(common.GetExeBaseName() + ".log")
	pLogger.Log("**********CLIENT程序启动**************")
	pLogger.Log("程序路径:", filepath.Join(common.GetExePath(), common.GetExeName()))
	pLogger.Log("当前时间:", common.GetNowTime())
	pLogger.Log("WINDIR:", common.GetWindowsDir())
	err := loadConfig()
	if err != nil {
		pLogger.LogExit("读取配置文件错误:", err)
		return
	}
	err = loadClientOnlyId()
	if err != nil {
		pLogger.LogExit("读取Agent唯一编号错误:", err)
		return
	}
	pLogger.Log("客户端标识:" + globalOnlyId)
	pLogger.Log("服务器IP:", serverHost)
	pLogger.Log("服务器HTTP端口:", serverHttpPort)
	pLogger.Log("服务器TCP端口:", serverTCPPort)
	pLogger.Log("服务器标识:", globalSvrFlag)
	pLogger.Log("进程ID:", globalProcessId)
}

//获得该客户端的唯一Id
func loadClientOnlyId() error {
	if globalOnlyId != "" {
		return nil
	}
	windir := common.GetWindowsDir()
	if windir == "" {
		panic("无法获取系统windows目录")
	}
	//如果windows目录下有onlyid就用Windows目录中的
	selfIdFile := "./17vinsoft_onlyid.dat"
	winIdFile := filepath.Join(windir, "17vinsoft_onlyid.dat")
	data, err := ioutil.ReadFile(winIdFile)
	if err == nil && len(data) > 0 { //就用这个,删除当前目录下的
		os.Remove(selfIdFile)
		globalOnlyId = string(data)
		return nil
	}
	//尝试使用当前目录下的
	data, err = ioutil.ReadFile(selfIdFile)
	if err == nil && len(data) > 0 {
		//尝试写入到windows目录,失败了也没关系,可能没有权限
		ioutil.WriteFile(winIdFile, data, 644)
		globalOnlyId = string(data)
		return nil
	}
	//到这里可能是首次运行程序,还没有生成过，那么就生成一个
	globalOnlyId = time.Now().Format("20060102150405.000")
	if ioutil.WriteFile(winIdFile, []byte(globalOnlyId), 644) != nil { //没有权限,写入当前
		if ioutil.WriteFile(selfIdFile, []byte(globalOnlyId), 644) != nil {
			return fmt.Errorf("无法写入globalOnlyId")
		}
	}
	return nil
}

//载入读取配置信息
func loadConfig() error {
	var err error
	mapConfig, err = common.LoadIniFile(".\\client_config.ini")
	if err != nil {
		return err
	}
	serverAddr = mapConfig["serveraddr"]
	if serverAddr == "" {
		return fmt.Errorf("配置服务器地址(serveraddr)不能为空,IP:Port的样式")
	}
	serverHost, serverHttpPort, err = net.SplitHostPort(serverAddr)
	if err != nil {
		return fmt.Errorf("解析配置地址:%s 错误:%v", serverAddr, err)
	}
	if n, err := strconv.Atoi(serverHttpPort); err == nil {
		if n > 0 && n < 65534 {
			serverTCPPort = strconv.Itoa(n + 1)
		}
	}
	globalSvrFlag = mapConfig["serverflag"]
	if globalSvrFlag == "" {
		return fmt.Errorf("配置服务器标识(serverflag)不能为空")
	}
	return nil
}
func getConfig(name string) string {
	return mapConfig[name]
}
