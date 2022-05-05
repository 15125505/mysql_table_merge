package models

import (
	"database/sql"
	"github.com/15125505/zlog/log"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

// DbInfo 数据库节点
type DbInfo struct {
	Name         string `json:"name"`
	CreateString string `json:"createString"`
	Tables       []struct {
		Name         string `json:"name"`
		CreateString string `json:"createString"`
	} `json:"tables"`
}
type DbInfos []DbInfo

// 获取该库节点下指定表格的创建字符串
func (dbInfo DbInfo) getTableCreateString(tableName string) string {
	for _, v := range dbInfo.Tables {
		if v.Name == tableName {
			return v.CreateString
		}
	}
	return ""
}

// 检查是否包含指定数据库
func (s DbInfos) contain(dbName string) bool {
	return s.find(dbName) != nil
}

// 根据数据库名称，查找所在的位置
func (s DbInfos) find(dbName string) *DbInfo {
	for _, v := range s {
		if v.Name == dbName {
			return &v
		}
	}
	return nil
}

// 根据库名和表名，查找对应的描述字符串
func (s DbInfos) getCreateString(dbName, tableName string) string {
	dbInfo := s.find(dbName)
	if dbInfo == nil {
		return ""
	}
	for _, v := range dbInfo.Tables {
		if v.Name == tableName {
			return v.CreateString
		}
	}
	return ""
}

// 获取指定的数据库连接
func getDbByString(dbString string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dbString)
	if err != nil {
		log.Error("打开数据库失败：", err.Error())
		return nil, err
	}
	db.SetConnMaxLifetime(time.Hour)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)
	return db, nil
}

// 检查字符串数组中是否存在某个字符串
func isStringInSlice(vec []string, str string) bool {
	for _, v := range vec {
		if v == str {
			return true
		}
	}
	return false
}

// TableSubItem 表格详细子项
type TableSubItem struct {
	Name string // 字段名称
	Sql  string // 字段的创建sql
}

// TableInfo 表格信息
type TableInfo struct {
	Name    string         // 表格名称
	Sql     string         // 表格的创建属性
	Columns []TableSubItem // 字段
	Keys    []TableSubItem // 索引
}

// 获取到指定字段的创建语句
func (t TableInfo) getColumnSql(name string) string {
	for _, v := range t.Columns {
		if v.Name == name {
			return v.Sql
		}
	}
	return ""
}

// 获取指定索引的创建语句
func (t TableInfo) getKeySql(name string) string {
	for _, v := range t.Keys {
		if v.Name == name {
			return v.Sql
		}
	}
	return ""
}
