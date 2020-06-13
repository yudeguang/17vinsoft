package common

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//默认的日志记录文件大小
var nDefaultLogMaxSize = int(50 * 1024*1024)

//日志记录类实现,face.FaceLogger实现
type BasicLogger struct {
	logFile    string        //日志文件路径
	maxSize    int64         //日志文件最大
	muxFile    sync.Mutex    //文件写入锁
	timeLayout string        //time时间样式
	doOnce     sync.Once     //执行一次的锁
	outBuffer  *bytes.Buffer //日志缓存
	flushTimer *time.Timer   //间隔写入日志的timer
}

func NewLogger(logFile string) *BasicLogger {
	obj := &BasicLogger{}
	obj.SetLogFile(logFile)
	return obj
}

func (b *BasicLogger) onceInit() {
	//没有设置日志文件
	if b.logFile == "" {
		panic("没有为日志对象设置文件名")
	}
	b.timeLayout = "01-02 15:04:05.000"
	if b.maxSize <= 0 {
		b.maxSize = int64(nDefaultLogMaxSize)
	}
	//创建日志缓存buffer
	b.outBuffer = bytes.NewBuffer(nil)
	b.flushTimer = time.AfterFunc(time.Second*3, b.base_on_timer)
}

//获得当前路径
func (b *BasicLogger) getExePath() string {
	path, err := os.Executable()
	if err != nil {
		return "./"
	}
	d, _ := filepath.Split(path)
	return d
}

//设置日志文件和大小,后来大小参数不用，用程序根据名字来判断
func (b *BasicLogger) SetLogFile(file string) {
	/*if strings.HasSuffix(strings.ToLower(file),".exe"){
		//file =
	}*/
	fileName := strings.ToLower(filepath.Base(file))
	b.logFile = filepath.Join(b.getExePath(), fileName)
	b.maxSize = int64(nDefaultLogMaxSize)
}

func (b *BasicLogger) Log(args ...interface{}) error {
	//执行一次初始化
	b.doOnce.Do(b.onceInit)
	//这里为了速度快一点,自己判断一下参数,因为大多数时候,Log的信息就只有一个字符串
	var infoText string = time.Now().Format(b.timeLayout)
	if len(args) == 1 {
		if s, ok := args[0].(string); ok {
			infoText += "@" + s + "\r\n"
		} else {
			infoText += "@" + fmt.Sprint(args...) + "\r\n"
		}
	} else {
		infoText += "@" + fmt.Sprint(args...) + "\r\n"
	}
	//同时输出屏幕和记录日志
	fmt.Print(infoText);
	//记录日志到缓存中定期刷新到日志
	b.muxFile.Lock()
	_, err := b.outBuffer.WriteString(infoText)
	if err != nil { //有问题,重置缓存
		b.outBuffer = bytes.NewBuffer(nil)
	}
	if b.outBuffer.Len() > 128*1024 { //大于128k就刷新
		b.base_flush()
	}
	b.muxFile.Unlock()
	return nil
}

func (b *BasicLogger) base_on_timer() {
	b.muxFile.Lock()
	defer b.muxFile.Unlock()
	if b.outBuffer.Len() > 0 {
		b.base_flush()
	}
	b.flushTimer.Reset(time.Second * 3)
}
//记录并退出程序
func (b *BasicLogger) LogExit(args ...interface{}){
	b.Log(args...)
	b.muxFile.Lock()
	if b.outBuffer.Len() > 0 {
		b.base_flush()
	}
	b.muxFile.Unlock()
	time.Sleep(time.Second)
	os.Exit(0)
}
//刷新数据到日志文件
func (b *BasicLogger) base_flush() {
	//退出时截断清空数据
	defer func() {
		b.outBuffer.Truncate(0)
	}()
	//打开文件,确定文件大小并写入文件
	isNewFile := false
	fsstat, err := os.Stat(b.logFile)
	if err == nil {
		if fsstat.Size() > b.maxSize {
			oldfile := b.logFile + ".old"
			os.Remove(oldfile)
			os.Rename(b.logFile, oldfile)
		}
	} else if os.IsNotExist(err) {
		isNewFile = true
	} else { //不知道什么原因了,忽略
		return
	}
	//打开文件并写入
	fs, err := os.OpenFile(b.logFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer fs.Close()
	if isNewFile {
		if _, err = fs.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
			return
		}
	}
	_, err = b.outBuffer.WriteTo(fs)
	fs.Close()
	return
}
