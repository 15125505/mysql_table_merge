package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/15125505/zlog/log"
)

// ImportDb sql导入到新的数据库
func ImportDb(dbString string, jsonBin []byte) (err error) {

	// 进行导入检查
	if err = ImportCheck(dbString, jsonBin); err != nil {
		log.Error("数据导入检查没有通过，取消导入！")
		return err
	}

	// 获取数据库连接
	db, err := getDbByString(dbString)
	if err != nil {
		log.Error("获取数据库连接失败")
		return
	}

	// 解析输入的数据
	vecNew := make(DbInfos, 0)
	err = json.Unmarshal(jsonBin, &vecNew)
	if err != nil {
		log.Error("解析处理的json内容失败：", err)
		return err
	}

	// 获取当前的数据
	vecOld, err := exportDb(db)
	if err != nil {
		return
	}

	// 基于导入的数据，进行处理
	for _, v := range vecNew {

		// 如果在旧库中没有的，那么进行创建库的行为
		oldDbInfo := vecOld.find(v.Name)
		if oldDbInfo == nil {
			log.Info("创建数据库：", v.Name)
			if err = exeSql(db, v.CreateString); err != nil {
				return
			}
			oldDbInfo = &DbInfo{
				Name:         v.Name,
				CreateString: v.CreateString,
				Tables:       nil,
			}
		}

		// 逐个表进行检查
		if err = exeSql(db, fmt.Sprintf("use `%v`", v.Name)); err != nil {
			return
		}
		for _, vv := range v.Tables {
			err = importTable(db, v.Name+"."+vv.Name, vv.CreateString, oldDbInfo.getTableCreateString(vv.Name))
			if err != nil {
				return
			}
		}
	}

	// 检查是否存在被修改的数据库（数据库的创建语句不同）
	for _, v := range vecOld {

		// 检查是否存在被删除的数据库（旧库中有，而新库中没有）
		dbInfoNew := vecNew.find(v.Name)
		if dbInfoNew == nil {
			err = errors.New("发现被删除的数据库：" + v.Name)
			log.Error(err)
			return
		}

		// 检查数据库是否被修改
		if isDbEqual(dbInfoNew.CreateString, v.Name) {
			err = errors.New("数据库发生变化：" + v.Name)
			log.Error(err)
			return
		}

		// 逐个表，检查是否已经被修改
		for _, vv := range v.Tables {

			// 找到新表中对应的创建字符串
			newTableString := vecNew.getCreateString(v.Name, vv.Name)
			if len(newTableString) == 0 {
				err = errors.New("发现被删除的数据表：" + v.Name + " => " + vv.Name)
				log.Error(err)
				return
			}

			// 检查数据表是否允许更新
			if !isTableValid(newTableString, vv.CreateString) {
				err = errors.New(fmt.Sprintf(`%v.%v检查不通过!!!`, v.Name, vv.Name))
				log.Error(err)
				return
			}
			//log.Info(fmt.Sprintf(`%v.%v检查通过...`, v.Name, vv.Name))
		}
	}

	log.Notice("新的数据库表结构已经被成功应用到当前数据库！")
	return
}

// 导入表格
func importTable(db *sql.DB, tablePath, newTableSql, oldTableSql string) (err error) {

	// 如果旧表sql为空字符串，表示之前没有这个表，那么直接创建即可
	if len(oldTableSql) == 0 {
		log.Info("创建数据表：", tablePath)
		return exeSql(db, newTableSql)
	}

	// 解析新旧表
	newTable, err := Sql2Table(newTableSql)
	if err != nil {
		log.Error("解析旧表失败：", err)
		return
	}
	oldTable, err := Sql2Table(oldTableSql)
	if err != nil {
		log.Error("解析新表失败：", err)
		return
	}

	// 检查新增字段
	for _, v := range newTable.Columns {
		var oldSql = oldTable.getColumnSql(v.Name)
		if oldSql != "" {
			continue
		}
		log.Info("为表格", tablePath, "增加字段", v.Name)
		err = exeSql(db, fmt.Sprintf(`alter table %v add column  %v`, tablePath, v.Sql))
		if err != nil {
			return
		}
	}

	// 检查新增索引
	for _, v := range newTable.Keys {
		var oldSql = oldTable.getKeySql(v.Name)
		if oldSql != "" {
			continue
		}
		log.Info("为表格", tablePath, "增加索引", v.Name)
		err = exeSql(db, fmt.Sprintf(`alter table %v add %v`, tablePath, v.Sql))
		if err != nil {
			return
		}
	}
	return
}

func exeSql(db *sql.DB, sqlString string) (err error) {
	_, err = db.Exec(sqlString)
	if err != nil {
		log.Error("执行语句失败：", err, "语句：", sqlString)
	}
	return
}
