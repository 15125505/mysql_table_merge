package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/15125505/zlog/log"
	"regexp"
	"strings"
)

// ImportCheck sql导入检查
func ImportCheck(dbString string, jsonBin []byte) (err error) {

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
				err = errors.New("发现被删除的数据表：" + v.Name + "." + vv.Name)
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

	return
}

// 检查数据库的创建语句是否相同
func isDbEqual(sqlNew, sqlOld string) bool {
	return sqlNew == sqlOld
}

// 检查表格是否满足自动插入要求
func isTableValid(sqlNew, sqlOld string) (ret bool) {

	// 默认返回ture
	ret = true

	// 解析新旧表
	newTable, err := Sql2Table(sqlNew)
	if err != nil {
		log.Error("解析旧表失败：", err)
		return false
	}
	oldTable, err := Sql2Table(sqlOld)
	if err != nil {
		log.Error("解析新表失败：", err)
		return false
	}

	// 检查表格属性
	if newTable.Sql != oldTable.Sql {
		log.Error(fmt.Sprintf(`新旧数据表属性不一致。旧：<%v> 新：<%v>`, oldTable.Sql, newTable.Sql))
		ret = false
	}

	// 检查字段是否变化
	for _, v := range oldTable.Columns {
		var newSql = newTable.getColumnSql(v.Name)
		if newSql == "" {
			log.Error("旧表中发现在新表中被删除的字段：", v.Name)
			ret = false
		} else if newSql != v.Sql {
			log.Error(fmt.Sprintf("字段 %v 发生变化 (%v) => (%v)", v.Name, v.Sql, newSql))
			ret = false
		}
	}

	// 检查索引是否变化
	for _, v := range oldTable.Keys {
		var newSql = newTable.getKeySql(v.Name)
		if newSql == "" {
			log.Error("旧表中发现在新表中被删除的索引：", v.Name)
			ret = false
		} else if newSql != v.Sql {
			log.Error(fmt.Sprintf("索引 %v 发生变化 (%v) => (%v)", v.Name, v.Sql, newSql))
			ret = false
		}
	}

	return
}
//
//// ParseTableString 解析表格的sql字符串为按列存储的数组（第一行是表格描述信息）
//func ParseTableString(tableSql string) (vecTableDetail []string) {
//
//	// 按行切分
//	vec := strings.Split(tableSql, "\n")
//
//	// 第一行是table名称
//	if matched, err := regexp.MatchString(`^CREATE TABLE.*\($`, vec[0]); err != nil || !matched {
//		log.Error("表格创建语句首行不匹配")
//		return
//	}
//
//	// 提取最后一行的表格描述部分，并去掉AUTO_INCREMENT和COMMENT内容
//	desc := regexp.MustCompile(`[^) ]+.*`).FindString(vec[len(vec)-1])
//	desc = regexp.MustCompile(`AUTO_INCREMENT=\d+ `).ReplaceAllString(desc, "")
//	desc = regexp.MustCompile(` COMMENT='.*'`).ReplaceAllString(desc, "")
//	vecTableDetail = append(vecTableDetail, desc)
//
//	// 提取每一行内容
//	for i := 1; i < len(vec)-1; i++ {
//		desc = vec[i]
//		desc = regexp.MustCompile(` COMMENT '.*'`).ReplaceAllString(desc, "") // 去掉注释
//		desc = regexp.MustCompile(`,$`).ReplaceAllString(desc, "")            // 去掉末尾的逗号
//		vecTableDetail = append(vecTableDetail, desc)
//	}
//
//	//log.Info("解析完成", vecTableDetail)
//	return
//}

// Sql2Table 根据sql创建table节点
func Sql2Table(createSql string) (tableInfo TableInfo, err error) {

	// 按行切分
	vec := strings.Split(createSql, "\n")
	if len(vec) == 0 {
		err = errors.New("输入的sql为空字符串")
		log.Error(err)
		return
	}

	// 解析表格名称
	ret := regexp.MustCompile("CREATE TABLE `?([^`]+)`? \\(").FindStringSubmatch(vec[0])
	if ret == nil {
		err = errors.New("解析表格名称失败")
		log.Error(err)
		return
	}
	tableInfo.Name = ret[1]

	// 解析表格属性
	ret = regexp.MustCompile(`[^) ]+(.*)`).FindStringSubmatch(vec[len(vec)-1])
	if ret == nil {
		err = errors.New("解析表格属性失败")
		log.Error(err)
		return
	}
	tableInfo.Sql = ret[1]
	tableInfo.Sql = regexp.MustCompile(`AUTO_INCREMENT=\d+ `).ReplaceAllString(tableInfo.Sql, "") // 去掉自增字段
	tableInfo.Sql = regexp.MustCompile(` COMMENT='.*'`).ReplaceAllString(tableInfo.Sql, "")       // 去掉注释

	// 将剩下的行，解析为字段和索引
	var name string
	for i := 1; i < len(vec)-1; i++ {
		var desc = vec[i]
		desc = regexp.MustCompile(` COMMENT '.*'`).ReplaceAllString(desc, "") // 去掉注释
		desc = regexp.MustCompile(`,$`).ReplaceAllString(desc, "")            // 去掉末尾的逗号
		if strings.Contains(desc, " KEY ") {                                  // 索引
			ret = regexp.MustCompile("KEY[` (]*([^` )]*)").FindStringSubmatch(desc)
			if ret == nil {
				err = errors.New("解析表格索引名称失败：" + desc)
				log.Error(err)
				return
			}
			name = ret[1]
			tableInfo.Keys = append(tableInfo.Keys, TableSubItem{
				Name: name,
				Sql:  desc,
			})
		} else { // 字段
			name = regexp.MustCompile("[^` ]+").FindString(desc)
			if len(name) == 0 {
				err = errors.New("解析表格字段名称失败：" + desc)
				log.Error(err)
				return
			}
			tableInfo.Columns = append(tableInfo.Columns, TableSubItem{
				Name: name,
				Sql:  desc,
			})
		}
	}

	//indent, _ := json.MarshalIndent(tableInfo, "", "  ")
	//log.Debug(string(indent))

	return
}
