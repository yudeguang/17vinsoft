package common

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//用于读取ini格式的文件,key不区分大小写
type IniConfiger struct{
	mapData map[string]string
}
//加载.ini文件,如果空参数则加载目录下的所有.ini后缀文件
func (this* IniConfiger) LoadAllFile(files ...string) error{
	if len(files) == 0{
		lstfs,err := ioutil.ReadDir(".")
		if err != nil{
			return err;
		}
		for _,fs := range lstfs{
			if fs.IsDir(){
				continue;
			}
			if strings.HasSuffix(strings.ToLower(fs.Name()),".ini"){
				files = append(files,fs.Name());
			}
		}
	}
	if len(files) == 0{
		return fmt.Errorf("can not find any ini file")
	}
	for _,f := range files{
		this.LoadFile(f);
	}
	return nil;
}
//加载一个文件
func (this* IniConfiger) LoadFile(file string) error{
	if this.mapData == nil{
		this.mapData = make(map[string]string)
	}
	data,err := ioutil.ReadFile(file);
	if err != nil{
		return err;
	}
	if bytes.HasPrefix(data,UTF8BOM){
		data = data[len(UTF8BOM):];
	}
	lines := strings.Split(string(data),"\n")
	for _, line := range lines {
		line = strings.TrimSpace(line);
		if !strings.HasPrefix(line, "#") { //#开头的认为是注释
			if npos := strings.Index(line, "="); npos > 0 {
				//解析变量时的左右两边空格忽略
				key := strings.TrimSpace(line[:npos])
				val := strings.TrimSpace(line[npos+1:])
				this.Set(key, val)
			}
		}
	}
	return nil;
}

//设置变量
func (this* IniConfiger) Set(key,val string){
	key = strings.ToLower(key)
	this.mapData[key] = val;
}
//重置
func (this* IniConfiger) Reset(){
	this.mapData = make(map[string]string)
}
//输出字符串
func (this* IniConfiger) ToString() string{
	if this.mapData == nil{
		return "";
	}
	out := bytes.NewBuffer(nil)
	for key, val := range this.mapData {
		out.WriteString(fmt.Sprintf("%s=%s\r\n",key,val))
	}
	return out.String()
}
func (this *IniConfiger) GetString(key string) string {
	if this.mapData == nil{
		return "";
	}
	if v,ok := this.mapData[strings.ToLower(key)];ok{
		return v;
	}
	return "";
}
//获取int
func (this *IniConfiger) GetInt64(key string) int64 {
	res := int64(0)
	if v := this.GetString(key);v != ""{
		res, _ = strconv.ParseInt(v, 10, 64)
	}
	return res
}
func (this *IniConfiger) GetInt(key string) int {
	return int(this.GetInt64(key))
}

func (this* IniConfiger) Exists(key string) bool{
	if(this.mapData == nil){
		return false;
	}
	key = strings.ToLower(key);
	_,ok := this.mapData[key]
	return ok;
}
func (this *IniConfiger) GetProxyDataDir() (string,error){
	key := "proxy_data_dir";
	dir := strings.TrimSpace(this.GetString(key))
	if !this.Exists(key){
		return "",fmt.Errorf("没有配置参数:%v 或为空",key);
	}
	fs,err := os.Stat(dir)
	if err != nil{
		return "",err;
	}
	if !fs.IsDir(){
		return "",errors.New(dir+" 不是目录")
	}
	return dir,nil;
}
//删除某个站点下的数据
/*
func (this *IniConfiger) RemoveProxySiteFiles(sitename string) (error){
	lst,err := this.GetProxySiteFiles(sitename);
	if err != nil{
		if os.IsNotExist(err){
			return nil;
		}
		return err;
	}
	for _,file := range lst{
		os.Remove(file);
	}
	return nil;
}*/
//查询某个站点下的所有文件,按文件名排序
func (this *IniConfiger) GetProxySiteFiles(sitename string) ([]string,error){
	basedir,err := this.GetProxyDataDir();
	if err != nil{
		return nil,err;
	}
	var sitedir = filepath.Join(basedir,sitename);
	lstFs,err := ioutil.ReadDir(sitedir);
	if err != nil{
		return nil,err;
	}
	lstResult := []string{}
	for _,fs := range lstFs{
		if fs.IsDir(){
			continue;
		}
		lstResult = append(lstResult,filepath.Join(sitedir,fs.Name()));
	}
	sort.Strings(lstResult)
	return lstResult,nil;
}