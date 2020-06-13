package common

import "strings"

type STRMap struct{
	mapData map[string]string;
}
//拆字符串获得一个STRMap
func NewSTRMap(text string) *STRMap{
	m := &STRMap{}
	m.parse(text);
	return m;
}

func (this* STRMap) parse(text string){
	this.mapData = make(map[string]string);
	for _,line := range strings.Split(string(text),"\n"){
		line = strings.TrimSpace(strings.Trim(line,"\r"));
		if strings.HasPrefix(line,"#"){
			continue;
		}
		if npos:=strings.Index(line,"=");npos>0{
			this.mapData[strings.ToLower(line[:npos])] = line[npos+1:];
		}
	}
}
func (this* STRMap) Exists(name string) bool{
	name =strings.ToLower(name);
	_,ok := this.mapData[name];
	return ok;
}
func (this* STRMap) GetString(name string) string{
	name =strings.ToLower(name);
	return this.mapData[name];
}