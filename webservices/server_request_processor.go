package webservices

import (
	"fmt"
	"github.com/yudeguang/17vinsoft/agentcomm"
	"log"
	"reflect"
)

//请求处理类
type RequestProcessor struct {
	ReqHeader *agentcomm.AgentHeader //请求的协议头
	ReqText   string                 //请求的数据
}

func (this *RequestProcessor) SetParam(header *agentcomm.AgentHeader, text string) {
	this.ReqHeader = header
	this.ReqText = text
}

func (this *RequestProcessor) Run() (*agentcomm.AgentHeader, string) {
	//先拷贝一个返回头,准备返回数据
	var respText = ""
	respHeader := &agentcomm.AgentHeader{}
	*respHeader = *this.ReqHeader
	respHeader.CmdId = agentcomm.CMD_ERROR //先定义为返回错误
	respHeader.ReverseId = 0
	respHeader.DataLen = 0
	respHeader.Resv = 0
	//根据命令号调用函数
	methodName := "CMD_0X" + fmt.Sprintf("%.8X", 1)
	v := reflect.ValueOf(this)
	if v.IsValid() {
		f := v.MethodByName(methodName)
		if f.IsValid() {
			var res = f.Call(nil)
			if len(res) == 1 {
				//只有1个返回值，如果这个返回值是error,则要作为信息返回Agent
				v := res[0].Interface()
				if v != nil {
					if err, ok := v.(error); ok {
						//error类型,认为失败
						respText = err.Error()
					} else if s, ok := v.(string); ok {
						//字符串类型,认为成功
						respHeader.CmdId = agentcomm.CMD_SUCCESS
						respText = s
					} else {
						//其它的类型,都认为是成功
						respHeader.CmdId = agentcomm.CMD_SUCCESS
						respText = fmt.Sprint(v)
					}
				} else { //返回了个nil值，认为成功的,
					respHeader.CmdId = agentcomm.CMD_SUCCESS
					respText = ""
				}
			} else {
				//返回非一个参数,认为错误一样
			}
		}
	}
	return respHeader, respText
}

func (this *RequestProcessor) CMD_0X00000001() error {
	log.Println("调用到这里了")
	return fmt.Errorf("我有什么信息吗?")
}
