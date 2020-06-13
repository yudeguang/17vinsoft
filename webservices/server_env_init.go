package webservices

import (
	"github.com/yudeguang/17vinsoft/common"
	"os"
	"path/filepath"
)

var theServerFlag string = "mytest_server"

//日志对象
var pLogger *common.BasicLogger = nil

//初始化配置函数
func InitWebServiceEnv() {
	pLogger = common.NewLogger(common.GetExeBaseName() + ".log")
	pLogger.Log("**********SERVER程序启动**************")
	pLogger.Log("程序路径:", filepath.Join(common.GetExePath(), common.GetExeName()))
	pLogger.Log("当前时间:", common.GetNowTime())
	pLogger.Log("进程ID:", os.Getpid())
	//创建sqlite数据库文件
	var err = pDBHelper.DBOpen()
	if err != nil {
		pLogger.LogExit("连接数据库失败:", err)
	}
}
