package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/15125505/zlog/log"
)

// ExportDb 导出数据库
func ExportDb(dbString string) (retData []byte, err error) {

	// 获取数据库连接
	db, err := getDbByString(dbString)
	if err != nil {
		log.Error("获取数据库连接失败")
		return
	}

	// 获取数据库内容
	vec, err := exportDb(db)
	if err != nil {
		return
	}

	retData, _ = json.Marshal(vec)
	return
}

// 导出数据表
func exportTable(db *sql.DB, dbName, tableName string) (createString string, err error) {
	err = db.QueryRow(fmt.Sprintf("show create table %v.%v", dbName, tableName)).Scan(&tableName, &createString)
	if err != nil {
		log.Error("获取数据表的查询语句失败：", err)
		return
	}
	//log.Info("数据表创建：", tableName, createString)
	return
}

// 导出数据内容
func exportDb(db *sql.DB) (vec DbInfos, err error) {

	// 几个系统库
	mapSystem := make(map[string]bool)
	mapSystem["information_schema"] = true
	mapSystem["mysql"] = true
	mapSystem["performance_schema"] = true
	mapSystem["sys"] = true

	// 获取数据库列表
	rows, err := db.Query("show databases")
	if err != nil {
		log.Error("查看数据表失败：", err)
		return
	}
	var name string
	names := make([]string, 0)
	for rows.Next() {
		err = rows.Scan(&name)
		if mapSystem[name] {
			//log.Notice("跳过系统库：", name)
			continue
		}
		names = append(names, name)
	}
	_ = rows.Close()

	// 对于每个数据库，查看一下它的创建语句
	var dbCreateString string
	for _, v := range names {

		// 创建一个节点
		var dbInfo = DbInfo{
			Name: v,
		}

		// 查看数据库的创造语句
		//log.Info("开始查看数据库：", v)
		err = db.QueryRow(fmt.Sprintf("show create database %v", v)).Scan(&name, &dbCreateString)
		if err != nil {
			log.Error("查看数据库的创建语句失败：", err)
			return
		}
		//log.Info("数据库创建语句：", v, dbCreateString)
		dbInfo.CreateString = dbCreateString

		// 获取该库下的所有表
		//log.Info("查看表结构：", v)
		var tableName string
		var tableNames = make([]string, 0)
		rows, err = db.Query("show tables from " + v)
		for rows.Next() {
			err = rows.Scan(&tableName)
			if err != nil {
				log.Error("获取表名失败：", err)
				return
			}
			tableNames = append(tableNames, tableName)
			//log.Info(fmt.Sprintf(`发现表%v.%v`, v, tableName))
		}
		_ = rows.Close()

		// 依次处理各个表
		for _, vv := range tableNames {
			dbCreateString, err = exportTable(db, name, vv)
			if err != nil {
				return
			}
			dbInfo.Tables = append(dbInfo.Tables, struct {
				Name         string `json:"name"`
				CreateString string `json:"createString"`
			}{Name: vv, CreateString: dbCreateString})
		}

		// 将这个数据库节点加入到总表中
		vec = append(vec, dbInfo)
	}

	return
}
