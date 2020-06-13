package webservices

import (
	"fmt"
	"github.com/astaxie/beego"
)

func registBeegoFuncMap() {
	beego.AddFuncMap("IsEqual", IsEqual)
	beego.AddFuncMap("NotEqual", NotEqual)
}
//比较两个类型是否相等，全部转换为字符串比较
func IsEqual(s1 interface{}, s2 interface{}) bool {
	return fmt.Sprint(s1) == fmt.Sprint(s2)
}
func NotEqual(s1 interface{}, s2 interface{}) bool {
	return fmt.Sprint(s1) != fmt.Sprint(s2)
}