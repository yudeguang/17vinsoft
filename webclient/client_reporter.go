package webclient

/*
import (
	"17vinsoft/common"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type clientReporter struct{
	agentInfo *common.TagAgentInfo
	replyInfo *common.TagAgentInfo
}
func (this* clientReporter) Init(){
	this.agentInfo = &common.TagAgentInfo{}
	this.agentInfo.OnlyId = globalOnlyId
	this.agentInfo.SeverFlag = globalSvrFlag;
	this.agentInfo.StartTime = common.StartTime;
	this.replyInfo = &common.TagAgentInfo{}
}

func (this *clientReporter) Start(onlyonce bool){
	for{
		if(!onlyonce){
			time.Sleep(time.Second*10)
		}
		err:=this.doReport();
		if err != nil{
			if(onlyonce){
				common.ExitProcess("clientReporter首次上报失败:",err)
			}else{
				common.ExitProcess("clientReporter循环上报失败:",err)
			}
		}
		if(onlyonce){ //首次只执行一次就可以
			break;
		}
	}
}

//上报数据和返回数据
func (this *clientReporter) doReport()(error){
	//log.Println("上报Agent信息...")
	data,err := json.Marshal(this.agentInfo);
	if err != nil{
		return err;
	}
	var sendUrl = fmt.Sprintf("http://%s/agent/updateself",serverAddr)
	resp,err := http.Post(sendUrl,"application/json",bytes.NewBuffer(data))
	if err != nil{
		return err;
	}
	defer resp.Body.Close();
	data,err = ioutil.ReadAll(resp.Body)
	if err != nil{
		return err;
	}
	err = json.Unmarshal(data,this.replyInfo)
	if err != nil{
		return err;
	}
	//log.Println("返回数据:",this.replyInfo.String())
	return nil;
}
*/