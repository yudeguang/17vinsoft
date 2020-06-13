package webclient

//间隔上报类
//var reporter = &clientReporter{}
//工作任务类
var worker = &clientWorker{};

//开始工作主函数
func StartWork(){
	//首先上报自己的信息并且获取配置，如果中间有通讯失败就退出，由服务再次加载起来
	//reporter.Init();
	worker.Init();
	err := pModuleManager.loadModuleList();
	if err != nil{
		pLogger.Log("读取支持模块错误:"+err.Error());
	}
	//开始工作,连接
	worker.Start();
}

