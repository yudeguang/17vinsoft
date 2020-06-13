package webclient

import (
	"github.com/yudeguang/17vinsoft/agentcomm"
)

func createSwitchData() *agentcomm.TagSwitchData {
	switchData := &agentcomm.TagSwitchData{}
	switchData.ProcId = globalProcessId
	switchData.OnlyId = globalOnlyId
	switchData.SvrFlag = globalSvrFlag
	switchData.StartTime = globalStartTime
	return switchData
}
