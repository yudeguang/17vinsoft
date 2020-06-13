package agentcomm
import (
	"bytes"
	"encoding/json"
	"strings"
)

type TagKVPair struct{
	Key string;
	Val interface{};
}
//一个模块条目
type TagModuleItem struct{
	Name string 	//名称
	ExePath string //可执行程序,全路径
}
type TagSwitchData struct{
	OnlyId 		string 			//客户端唯一id
	SvrFlag 	string 			//服务端标记号,服务端需要验证是否标记号一致，一致才处理请求
	StartTime 	string 			//程序启动时间,用于判断错误
	ProcId 		string 			//进程id
	RemoteAddr 	string 			//路由的IP:PORT地址
	HostAddr 	string 			//本机真是的IP:PORT地址
	Modules 	[]*TagModuleItem //支持的模块信息
}
//生成格式化的字符串
func (this *TagSwitchData) ToString() string{
	data,err := json.MarshalIndent(this,"","\t");
	if err != nil{
		return "";
	}
	return string(data);
}
//与json之间的转换
func SwitchDataFromJson(s string)(*TagSwitchData,error){
	item := &TagSwitchData{}
	err := json.Unmarshal([]byte(s),item)
	if err != nil{
		return nil,err;
	}
	return item,nil;
}
func SwitchDataToJson(p *TagSwitchData) string{
	if(p == nil){
		return ""
	}
	data,err := json.Marshal(p);
	if err != nil{
		return "";
	}
	return string(data);
}
//与Agent交互任务的参数
type TagTaskParam struct{
	TaskName 		string 	//任务名,对应的服务器端ModuleName
	TimeoutSecond 	int 	//最大执行超时时间
	LstLine 		[]string //变量Key与value;
}

func (this* TagTaskParam) ToIniString() string{
	out := bytes.NewBuffer(nil);
	out.WriteString("##auto create by agent server,do not modify\r\n")
	for _,line := range this.LstLine{
		line = strings.TrimSpace(line);
		out.WriteString(line+"\r\n");
	}
	return out.String();
}