package webservices

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

type AgentController struct {
	BaseController
}
func (this *AgentController) ListAgent(){
	lstAgent,err := pDBHelper.GetAgentList(0);
	if err != nil{
		this.Ctx.WriteString("查询Agent列表失败:"+err.Error())
		return;
	}
	this.Data["AgentList"] = lstAgent;
	//log.Println(common.ToJsonString(lstAgent))
	this.TplName = "page_agent_list.html";
}
//删除一个agent信息,删除后会重新再上来
func (this *AgentController) DeleteAgent(){
	id,err := this.GetInt("id");
	if err != nil{
		this.replyERRJson("参数错误:"+err.Error())
		return;
	}
	err = pDBHelper.DeleteAgent(id);
	if err != nil{
		this.replyERRJson("删除失败:"+err.Error())
	}else{
		this.replyOKJson("删除成功")
	}
}
func (this* AgentController) EditConfig(){
	start :=time.Now()
	defer func(){
		log.Println("编辑耗时:",time.Now().Sub(start))
	}();
	id,err := this.GetInt("id");
	if err != nil{
		this.Ctx.WriteString ("参数错误:"+err.Error())
		return;
	}
	record,err := pDBHelper.GetAgentRecordById(id)
	if err != nil{
		this.Ctx.WriteString("查询Agent信息错误:"+err.Error());
		return;
	}
	modules,err := pDBHelper.GetModuleByOnlyId(record.OnlyId)
	if err != nil{
		this.Ctx.WriteString("查询Module信息错误:"+err.Error());
		return;
	}
	this.Data["record"] = record;
	this.Data["modules"] = modules;
	this.Data["prioritylist"] = []int{0,1,2,3,4,5,6,7,8,9}
	this.Data["id"] = id;
	this.TplName="page_edit_config.html"
}

func (this* AgentController) SaveConfig(){
	Id,err := this.GetInt("Id");
	if err != nil || Id<=0{
		this.replyERRJson("Id参数错误")
		return;
	}
	ChsName:=this.GetString("ChsName");
	Priority,err := this.GetInt("Priority")
	if err != nil || Id<=0{
		this.replyERRJson("Priority参数错误")
		return;
	}
	//获得模块启停信息
	lstModuleId := this.GetStrings("ModuleId")
	lstModuleName := this.GetStrings("ModuleName")
	lstIsValid := this.GetStrings("IsValid")
	lstMaxSecond := this.GetStrings("MaxSecond")
	lstStaticVar := this.GetStrings("StaticVar")
	nSize := len(lstModuleId);
	if len(lstModuleName) != nSize || len(lstIsValid) != nSize || len(lstMaxSecond) != nSize{
		this.replyERRJson("模块配置信息长度不正确")
		return;
	}
	//检查lstModuleId是否都为数字
	for idx,mid := range lstModuleId{
		n,err := strconv.Atoi(mid);
		if err != nil || n<=0{
			this.replyERRJson(fmt.Sprintf("ModuleId参数错误:%v",lstModuleId));
			return;
		}
		if lstIsValid[idx] != "1"{
			lstIsValid[idx] = "0";
		}
		if lstIsValid[idx]=="1"{
			n,err = strconv.Atoi(lstMaxSecond[idx]);
			if err != nil || n<=0{
				this.replyERRJson(fmt.Sprintf("模块:%s 等待时间必须为大于0的数字",lstModuleName[idx]))
				return;
			}
		}else{
			lstMaxSecond[idx] = "30"
		}
	}
	//先保存Agent基本信息
	err = pDBHelper.Exec("UPDATE AgentList SET ChsName=?,Priority=? WHERE Id=?",ChsName,Priority,Id);
	if err != nil{
		this.replyERRJson("保存Agent信息错误:"+err.Error())
		return;
	}
	var hasReplyed = false;
	for idx,mid := range lstModuleId{
		err = pDBHelper.Exec("UPDATE AgentModule SET IsValid=?,Timeout=?,StaticVar=? WHERE Id=?",
			lstIsValid[idx],lstMaxSecond[idx],lstStaticVar[idx],mid);
		if err != nil && !hasReplyed{
			this.replyERRJson("保存Module信息错误:"+err.Error())
			hasReplyed = true;
		}
	}
	if(!hasReplyed){
		this.replyOKJson("保存成功")
	}
}