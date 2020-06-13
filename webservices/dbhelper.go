package webservices

import (
	"database/sql"
	"fmt"
	"github.com/yudeguang/17vinsoft/agentcomm"
	"github.com/yudeguang/17vinsoft/common"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var pDBHelper = &clsDBHelper{}

func ReplaceSQLChar(sour string) string {
	sour = strings.Replace(sour, "'", "''", -1)
	sour = strings.Replace(sour, "\\", "\\\\", -1)
	return sour
}

//数据库辅助类
type clsDBHelper struct {
	pSqliteDB *sql.DB
}

//定义需要创建表的sql语句
var lstSqlText = []string{
	`CREATE TABLE IF NOT EXISTS AgentList(
		"Id"  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
		"OnlyId"  	 VARCHAR(200) NOT NULL,
		"ChsName" 	 VARCHAR(200) DEFAULT '',
		"Priority" 	 INT DEFAULT 0,
		"AgentAddr"  VARCHAR(50) DEFAULT '',
		"ProcessId"  VARCHAR(20) DEFAULT '',
		"FindTime"   VARCHAR(50) DEFAULT '',
		"StartTime"  VARCHAR(50) DEFAULT '',
		"ReportTime" VARCHAR(50) DEFAULT ''
		)`,
	`CREATE TABLE IF NOT EXISTS AgentModule(
		"Id"  		INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
		"OnlyId"  		VARCHAR(200) NOT NULL,
		"Name"  		VARCHAR(1024) DEFAULT '',
		"IsValid"  		INT DEFAULT 0,
		"Timeout" 		INT DEFAULT 30,
		"StaticVar"		VARCHAR(2048) DEFAULT ''
		)`,
	`CREATE VIEW IF NOT EXISTS V_ModuleAll AS
SELECT OnlyId,ifnull(group_concat(Name, ','),'') AS Modules FROM AgentModule GROUP BY OnlyId
`,
	`CREATE VIEW IF NOT EXISTS V_ModuleValid AS
SELECT OnlyId,ifnull(group_concat(Name, ','),'') AS Modules FROM AgentModule WHERE ifnull(IsValid, 0) = 1 GROUP BY OnlyId
`,
}

//定义AgentList表结构
type TagAgentInfoRecord struct {
	Id         int
	OnlyId     string
	ChsName    string
	Priority   int
	AgentAddr  string
	ProcessId  string
	FindTime   string
	StartTime  string
	ReportTime string
	Resv1      interface{} //用来向页面传递参数
}

//AgentModule结构
type TagModuleInfoRecord struct {
	Id        int
	OnlyId    string
	Name      string
	IsValid   int
	Timeout   int
	StaticVar string //静态参数
}

//打开数据库连接
func (this *clsDBHelper) DBOpen() error {
	var err error
	this.pSqliteDB, err = sql.Open("sqlite3", "database.db3")
	if err != nil {
		return err
	}
	for _, text := range lstSqlText {
		_, err = this.pSqliteDB.Exec(text)
		if err != nil {
			return err
		}
	}
	return nil
}

//关闭数据库连接
func (this *clsDBHelper) DBClose() {
	if this.pSqliteDB != nil {
		this.pSqliteDB.Close()
		this.pSqliteDB = nil
	}
}

//获得连接
func (this *clsDBHelper) GetDB() *sql.DB {
	return this.pSqliteDB
}

//获取配置的值
func (this *clsDBHelper) GetConfigValue(keyName string) string {
	var value = ""
	var err = this.pSqliteDB.QueryRow("SELECT ifnull(KeyValue,'') FROM Config WHERE KeyName=?", keyName).Scan(&value)
	if err != nil && err != sql.ErrNoRows {
		log.Println("查询配置失败:", err)
		return ""
	}
	return value
}

//记录sql语句错误
func (this *clsDBHelper) logSQL(sqlText string, err error) {
	if err != nil {
		pLogger.Log("执行SQL语句错误:\r\n" + sqlText + "\r\n错误:" + err.Error())
	} else {
		pLogger.Log("执行SQL语句成功:\r\n" + sqlText)
	}
}

//直接执行某个SQL语句
func (this *clsDBHelper) Exec(sqlText string, args ...interface{}) error {
	_, err := this.pSqliteDB.Exec(sqlText, args...)
	if err != nil {
		this.logSQL(sqlText, err)
	}
	return err
}

//处理主连接的信息入库,这里是设备存活信息
func (this *clsDBHelper) UpdateAgentFromMainConnect(switchData *agentcomm.TagSwitchData, needUpdateModule bool) error {
	onlyid := switchData.OnlyId
	if onlyid == "" {
		return fmt.Errorf("param onlyid is empty")
	}
	var Id int = 0
	var sqlText = fmt.Sprintf("SELECT Id FROM AgentList WHERE OnlyId='%s'", ReplaceSQLChar(onlyid))
	var err = this.pSqliteDB.QueryRow(sqlText).Scan(&Id)
	if err != nil {
		if err != sql.ErrNoRows {
			this.logSQL(sqlText, err)
			//查询这个都错误，算了
			return err
		} else {
			Id = 0
		}
	}
	var nowTime = common.GetNowTime()
	var startTime = switchData.StartTime
	var processId = switchData.ProcId
	var addr = switchData.RemoteAddr
	if Id == 0 { //没有记录
		sqlText = fmt.Sprintf("INSERT INTO AgentList(OnlyId,AgentAddr,FindTime,StartTime,ReportTime,ProcessId) VALUES ('%s','%s','%s','%s','%s','%s')",
			ReplaceSQLChar(onlyid),
			addr,
			nowTime,
			startTime,
			nowTime,
			processId)
	} else {
		sqlText = "UPDATE AgentList SET ReportTime='" + nowTime + "'"
		if addr != "" {
			sqlText += ",AgentAddr='" + addr + "'"
		}
		if startTime != "" {
			sqlText += ",StartTime='" + startTime + "'"
		}
		if processId != "" {
			sqlText += ",ProcessId='" + processId + "'"
		}
		sqlText += " WHERE Id=" + fmt.Sprint(Id)
	}
	if _, err = this.pSqliteDB.Exec(sqlText); err != nil {
		this.logSQL(sqlText, err)
		return err
	}
	//只有主连接第一次连接需要更新模块
	if needUpdateModule {
		//删除已经不支持的模块,添加里面没有的模块
		sqlText = "SELECT ifnull(Modules,'') FROM V_ModuleAll WHERE OnlyId=?"
		var validModules = ""
		err = this.pSqliteDB.QueryRow(sqlText, onlyid).Scan(&validModules)
		if err != nil && err != sql.ErrNoRows {
			this.logSQL(sqlText, err)
			return err
		}
		//把已经存在的先放在mapExists里,后面如果报上来的已经存在了就删除
		var mapExists = make(map[string]int)
		for _, name := range strings.Split(validModules, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				mapExists[strings.ToLower(name)] = 1
			}
		}
		for _, module := range switchData.Modules {
			smallName := strings.ToLower(module.Name)
			if _, ok := mapExists[smallName]; ok {
				delete(mapExists, smallName)
				continue
			}
			//不存在,加入一个
			sqlText = "INSERT INTO AgentModule(OnlyId,Name) VALUES (?,?)"
			_, err = this.pSqliteDB.Exec(sqlText, onlyid, strings.ToLower(module.Name))
			if err != nil {
				this.logSQL(sqlText, err)
			}
		}
		//删除mapExists中海没被移除掉的
		for name, _ := range mapExists {
			sqlText = "DELETE FROM AgentModule WHERE OnlyId=? AND Name=?"
			_, err = this.pSqliteDB.Exec(sqlText, onlyid, name)
			if err != nil {
				this.logSQL(sqlText, err)
			}
		}
	}
	return nil
}

//查询已经存在的Agent列表,如果Id!=0,则返回所有
func (this *clsDBHelper) GetAgentList(Id int) ([]*TagAgentInfoRecord, error) {
	var lstAgent = []*TagAgentInfoRecord{}
	var sqlText = `SELECT A.Id,A.OnlyId,A.ChsName,A.Priority,A.AgentAddr,A.ProcessId,A.FindTime,A.StartTime,A.ReportTime,ifnull(B.Modules,'')
	FROM AgentList AS A LEFT JOIN V_ModuleValid AS B ON A.OnlyId=B.OnlyId`
	if Id != 0 {
		sqlText += fmt.Sprintf(" WHERE A.Id=%d", Id)
	}
	rows, err := this.pSqliteDB.Query(sqlText)
	if err != nil {
		return lstAgent, err
	}
	defer rows.Close()
	for rows.Next() {
		var modListString = ""
		item := &TagAgentInfoRecord{}
		err = rows.Scan(&item.Id,
			&item.OnlyId,
			&item.ChsName,
			&item.Priority,
			&item.AgentAddr,
			&item.ProcessId,
			&item.FindTime,
			&item.StartTime,
			&item.ReportTime,
			&modListString)
		if err != nil {
			return lstAgent, err
		}
		item.Resv1 = modListString
		lstAgent = append(lstAgent, item)
	}
	return lstAgent, err
}

//根据Id查询一条Agent记录
func (this *clsDBHelper) GetAgentRecordById(Id int) (*TagAgentInfoRecord, error) {
	lst, err := this.GetAgentList(Id)
	if err != nil {
		return nil, err
	}
	if len(lst) != 1 {
		return nil, fmt.Errorf("没有匹配的记录:Id=%v", Id)
	}
	return lst[0], nil
}

//删除一个条目
func (this *clsDBHelper) DeleteAgent(Id int) error {
	//先查出来OnlyId,根据OnlyId删除模块表
	var onlyId string = ""
	var err = this.pSqliteDB.QueryRow("SELECT OnlyId FROM AgentList WHERE Id=?", Id).Scan(&onlyId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	this.pSqliteDB.Exec("DELETE FROM AgentModule WHERE OnlyId=?", onlyId)
	_, err = this.pSqliteDB.Exec("DELETE FROM AgentList WHERE Id=?", Id)
	return err
}

//查询某个OnlyId包含的模块
func (this *clsDBHelper) GetModuleByOnlyId(onlyid string) ([]*TagModuleInfoRecord, error) {
	var sqlText = "SELECT Id,OnlyId,Name,IsValid,ifnull(Timeout,30),ifnull(StaticVar,'') FROM AgentModule"
	if onlyid != "" {
		sqlText += fmt.Sprintf(" WHERE OnlyId='" + onlyid + "'")
	}
	rows, err := this.pSqliteDB.Query(sqlText)
	if err != nil {
		this.logSQL(sqlText, err)
		return nil, err
	}
	defer rows.Close()
	lst := []*TagModuleInfoRecord{}
	for rows.Next() {
		item := &TagModuleInfoRecord{}
		err = rows.Scan(&item.Id,
			&item.OnlyId,
			&item.Name,
			&item.IsValid,
			&item.Timeout,
			&item.StaticVar)
		if err != nil {
			this.logSQL("循环结果集AgentModule", err)
			return nil, err
		}
		lst = append(lst, item)
	}
	return lst, nil
}
