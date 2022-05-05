package main

import (
	"flag"
	"fmt"
	"github.com/15125505/zlog/log"
	"io/ioutil"
	"mysql_table_merge/models"
)

func main() {

	// 自定义使用说明
	isUsageShown := false
	flag.Usage = func() {
		fmt.Println("使用说明：")
		fmt.Println("  示例：mysql_table_merge -db 'root:password@tcp(127.0.0.1:3306)/?charset=utf8mb4' -mode check -file dbExport.json")
		flag.PrintDefaults()
		isUsageShown = true
	}

	// 获取用户输入
	var db string
	var mode string
	var file string
	flag.StringVar(&db, "db", "", "数据库连接字符串，如：root:password@tcp(127.0.0.1:3306)/?charset=utf8mb4")
	flag.StringVar(&mode, "mode", "check", "工作模式，允许的值为(check,import,export)其中之一")
	flag.StringVar(&file, "file", "dbExport.json", "导入或导出的数据表结构文件路径")
	flag.Parse()

	// 如果没有获得db而且没有显示过使用说明，那么强制显示使用说明
	if db == "" && !isUsageShown {
		flag.Usage()
		return
	}

	// 执行用户操作
	switch mode {
	case "check":
		f, err := ioutil.ReadFile(file)
		if err != nil {
			log.Error("读取文件", file, "失败")
			return
		}
		err = models.ImportCheck(db, f)
		if err != nil {
			log.Error("检查没有通过：", err)
			return
		}
		log.Notice("经过检查：该数据库表结构可以被正常导入！")
	case "import":
		f, err := ioutil.ReadFile(file)
		if err != nil {
			log.Error("读取文件", file, "失败")
			return
		}
		err = models.ImportDb(db, f)
		if err != nil {
			log.Error("导入表结构失败：", err)
		}
	case "export":
		retData, err := models.ExportDb(db)
		if err != nil {
			return
		}
		err = ioutil.WriteFile(file, retData, 0664)
		if err != nil {
			log.Error("保存文件", file, "失败")
			return
		}
		log.Notice("数据库表结构已经被成功导出！")
	default:
		log.Error("mode参数不合法，必须为(check,import,export)其中之一")
	}
}
