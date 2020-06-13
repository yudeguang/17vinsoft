package webclient

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yudeguang/17vinsoft/agentcomm"
	"github.com/yudeguang/17vinsoft/common"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type clsClientSession struct {
	reverseId int64
	conn      net.Conn
	reqParam  *agentcomm.TagTaskParam
	reqHeader *agentcomm.AgentHeader
	reqText   string
	taskName  string
}

func (this *clsClientSession) logInfo(args ...interface{}) {
	info := fmt.Sprint(args...)
	pLogger.Log(fmt.Sprintf("[reverseId:%d]", this.reverseId), info)
}
func (this *clsClientSession) Run(reverseId int64) {
	var err error = nil
	defer func() {
		if this.conn != nil {
			this.conn.Close()
			this.conn = nil
		}
		if err != nil {
			this.logInfo(err)
		}
	}()
	this.reverseId = reverseId
	this.conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%s", serverHost, serverTCPPort), time.Second*5)
	if err != nil {
		return
	}
	this.conn.(*net.TCPConn).SetLinger(20)
	err = this.conn.SetDeadline(time.Now().Add(time.Second * 10))
	if err != nil {
		return
	}
	this.conn.(*net.TCPConn).SetKeepAlive(true)
	//发送我回复的ID
	err = agentcomm.WritePackage(this.conn, agentcomm.CMD_REPLY_REVERSE_CONNECT, "", this.reverseId)
	if err != nil {
		return
	}
	//接收要让我做的事情
	this.reqHeader, this.reqText, err = agentcomm.ReadPackage(this.conn, 60)
	if err != nil {
		return
	}
	replyText, err := this.executeTask()
	if err != nil {
		agentcomm.WritePackage(this.conn, agentcomm.CMD_ERROR, err.Error())
		time.Sleep(time.Second * 3)
		return
	}
	//这里成功了,发送数据
	err = agentcomm.WritePackage(this.conn, agentcomm.CMD_SUCCESS, replyText)
	if err != nil {
		pLogger.Log("发送任务结果数据失败:", err)
		return
	}
	//这里重新读取命令,保证数据发送完成
	agentcomm.ReadPackage(this.conn, 60)
	pLogger.Log("任务:" + this.taskName + " 请求/交互结束")
}
func (this *clsClientSession) executeTask() (string, error) {
	if this.reqHeader.CmdId != agentcomm.CMD_REQUEST_EXECUTE_TASK {
		return "", fmt.Errorf("不支持的命令类型")
	}
	this.reqParam = &agentcomm.TagTaskParam{}
	var err = json.Unmarshal([]byte(this.reqText), this.reqParam)
	if err != nil || this.reqParam.TaskName == "" {
		return "", fmt.Errorf("解析任务请求参数错误")
	}
	this.taskName = this.reqParam.TaskName
	moduleDir := filepath.Join("./module", this.taskName)
	configFile := filepath.Join(moduleDir, "dynamic_config.ini")
	exeFile := filepath.Join(common.GetExePath(), "module", this.taskName, this.taskName+".exe")
	if _, err := os.Stat(exeFile); err != nil {
		return "", err
	}
	if err = ioutil.WriteFile(configFile, []byte(this.reqParam.ToIniString()), 644); err != nil {
		return "", err
	}
	//删除结果文件后执行程序
	resultFile := filepath.Join(moduleDir, "result.txt")
	os.Remove(resultFile)
	err = this.executeProgram(exeFile)
	if err != nil {
		return "", fmt.Errorf("执行程序错误:%v", err)
	}
	//提取数据并返回
	data, err := ioutil.ReadFile(resultFile)
	if err != nil {
		return "", fmt.Errorf("没有生成结果文件:%v", resultFile)
	}
	return string(data), nil
}
func (this *clsClientSession) executeProgram(exeFile string) error {
	var timeoutSecond = this.reqParam.TimeoutSecond
	this.logInfo("执行程序:" + exeFile)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSecond))
	defer cancel()
	cmd := exec.CommandContext(ctx, exeFile, "127.0.0.1")
	cmd.Dir = filepath.Dir(exeFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = os.Stderr
	var err = cmd.Start()
	if err != nil {
		return err
	}
	cmd.Wait()
	err = nil
	select {
	case <-ctx.Done():
		err = fmt.Errorf("执行超时:%v s", timeoutSecond)
		this.logInfo(err)
		break
	default:
		this.logInfo("执行程序完成 ")
		break
	}
	return err
}
