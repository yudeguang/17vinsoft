package common

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//import "os"
var ErrDirNotIncludeDot = errors.New("RemoveDirAllFiles dir params must with dot(.)")
//删除某个目录下的所有文件,这个函数比较危险,害怕直接参数给了个c盘,那文件全部就删掉了
//在我们的应用里都是删除站点,站点肯定带有 "."字符,所以判断一下这个
func RemoveSiteAllFiles(dir string) error{
	if(!strings.Contains(dir,".")){
		return ErrDirNotIncludeDot
	}
	fs,err := os.Stat(dir)
	if err != nil{
		if os.IsNotExist(err){
			//目录本身就不存在，认为成功
			return nil;
		}
		return err;
	}
	if !fs.IsDir(){
		return fmt.Errorf("dir param not a directory")
	}
	err = filepath.Walk(dir,func(path string, info os.FileInfo, err error) error{
		if !info.IsDir(){
			return os.Remove(path)
		}
		return nil;
	})
	return err;
}
//压缩目录下的所有文件到一个zip文件
//dstZipFile:目标zip文件
//dirs:目录列表，必须要有一个
func CompressDirFileToZIP(dstZipFile string,dirs ...string) error{
	if len(dirs) == 0{
		return fmt.Errorf("必须至少压缩一个目录")
	}
	// 创建 zip 包文件
	zfile, err := os.Create(dstZipFile)
	if err != nil {
		return err;
	}
	defer zfile.Close()
	// 实例化新的 zip.Writer
	zw := zip.NewWriter(zfile)
	defer zw.Close()
	//获取各个目录下的文件压缩
	for _,dir := range dirs{
		lstFs,err := ioutil.ReadDir(dir);
		if err != nil{
			continue;
		}
		for _,fs := range lstFs{
			if fs.IsDir(){ //目录就不压缩了
				continue;
			}
			err = compressOneFileToZIP(filepath.Join(dir,fs.Name()),zw);
			if err != nil{
				log.Println("压缩文件错误:"+err.Error())
			}
		}
	}
	return nil;
}
func compressOneFileToZIP(file string,zw *zip.Writer) error{
	fr, err := os.Open(file)
	if err != nil {
		return err
	}
	fi, err := fr.Stat()
	if err != nil {
		return err
	}
	// 写入文件的头信息
	fh, err := zip.FileInfoHeader(fi)
	if err != nil{
		return err;
	}
	fh.Method = zip.Deflate
	w, err := zw.CreateHeader(fh)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, fr)
	if err != nil {
		return err
	}
	return nil;
}