package webclient

import (
	"github.com/yudeguang/17vinsoft/agentcomm"
	"io/ioutil"
	"os"
	"path/filepath"
)

var pModuleManager = &clsModuleManager{}
var modulePath = "./module"

//模块管理类
type clsModuleManager struct {
	lstModule []*agentcomm.TagModuleItem
}

//获得模块列表
func (this *clsModuleManager) GetModuleList() []*agentcomm.TagModuleItem {
	return this.lstModule
}

//读取一次模块列表
func (this *clsModuleManager) loadModuleList() error {
	this.lstModule = []*agentcomm.TagModuleItem{}
	lstFile, err := ioutil.ReadDir(modulePath)
	if err != nil {
		return err
	}
	//module的规则是/module/dirname/dirname.exe
	for _, fs := range lstFile {
		if !fs.IsDir() {
			continue
		}
		var dirName = fs.Name()
		var exeFile = filepath.Join(modulePath, dirName, dirName+".exe")
		if _, err = os.Stat(exeFile); err != nil {
			continue
		}
		item := &agentcomm.TagModuleItem{}
		item.Name = dirName
		item.ExePath = exeFile
		this.lstModule = append(this.lstModule, item)
	}
	return nil
}
