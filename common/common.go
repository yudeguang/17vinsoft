package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//常用的一些初始化,主要是设置环境目录
/*
var ExePath string = ""
var ExeFile string = ""
var BaseExeName string = "" //不带后缀的课执行名称
var StartTime string = ""
*/
//定义常用的变量
var exePath string = "" //当前exe文件路径
var exeName string = "" //当前exe文件的名字
var startTime string = "" //当前程序的启动时间
var winDir string = "" //windows目录,windows平台有效
var UTF8BOM = []byte{0xEF,0xBB,0xBF}
//http服务的端口
var ServerHttpPort = 9000
var ServerTCPPort = ServerHttpPort+1;
//初始化环境
func init(){
	s, err := os.Executable();
	if err != nil{
		panic(err);
	}
	exePath, exeName = filepath.Split(s)
	os.Chdir(exePath);
	startTime = time.Now().Format("2006-01-02 15:04:05")
}
//获得exe文件路径和文件名字
func GetExePath() (string){
	return exePath;
}
func GetExeName() (string){
	return exeName;
}
//获得不带后缀的exe名称，一般用于记录日志文件
func GetExeBaseName() string{
	pos := strings.LastIndex(exeName,".")
	if pos>0{
		return exeName[:pos]
	}
	return exeName;
}
//获得程序的启动时间
func GetStartTime() string{
	return startTime;
}

//获得系统windows目录
func GetWindowsDir() string{
	if winDir!=""{
		return winDir;
	}
	winDir = os.Getenv("windir")
	if(winDir == ""){
		for c := 'c';c<'z';c++{
			tdir := fmt.Sprintf("%s:\\windows",string(c));
			if fs,err := os.Stat(tdir);err == nil && fs.IsDir(){
				winDir = tdir;
				break;
			}
		}
	}
	return winDir;
}

//将.ini文件格式的配置转为map,所有的key用小写保存
func LoadIniFile(file string) (map[string]string,error){
	data,err := ioutil.ReadFile(file);
	if err != nil{
		return nil,err;
	}
	var m = make(map[string]string);
	for _,line := range strings.Split(string(data),"\n"){
		line = strings.TrimSpace(strings.Trim(line,"\r"));
		if strings.HasPrefix(line,"#"){
			continue;
		}
		if npos:=strings.Index(line,"=");npos>0{
			m[strings.ToLower(line[:npos])] = line[npos+1:];
		}
	}
	return m,nil;
}

//退出程序
func ExitProcess(args... interface{}){
	info := fmt.Sprint(args...);
	log.Println("退出程序,原因:"+info);
	time.Sleep(1*time.Second)
	os.Exit(0);
}

//获得当前系统时间
func GetNowTime() string{
	return time.Now().Format("2006-01-02 15:04:05")
}
//对象转为json
func ToJsonString(o interface{}) string{
	if o == nil{
		return "nil";
	}
	data,err := json.MarshalIndent(o,"","\t")
	if err != nil{
		return "ERR:"+err.Error();
	}
	return string(data);
}

//可能是相对路径转为绝对路径
func ToAbsPath(s string) string{
	if filepath.IsAbs(s){
		return s;
	}
	dst,err := filepath.Abs(s);
	if err != nil{
		return s;
	}
	return dst;
}