# mysql表结构自动同步工具

* 本工具用于使用一个数据库（例如开发数据库）的表结构去自动同步另外一个数据库（例如线上数据库）的表结构。
* 最典型的使用场景是：后台服务开发过程中，如果如果表结构有变更，在开发测试完成之后，需要将表结构的更新同步到线上数据。使用这个工具，可以避免手动的错误或疏漏。

## 一. 重要约定

为了保证数据库对服务端代码的前项兼容，也为了降低数据表结构同步的复杂度，特做出如下约定（假如要将A数据库的表结构，同步到B数据库）：

* 数据库A相对于数据库B，只允许增加数据库，不允许删除或修改数据库。
* 数据库A相对于数据库B，只允许增加数据表，不允许删除或修改数据表属性（此处属性不包括数据表字段，另外注释和AUTO_INCREMENT例外）。
* 数据库A相对于数据库B，只允许增加数据表字段，不允许删除或修改数据表字段（注释的变更例外）。

## 二. 安装

go install github.com/15125505/mysql_table_merge@latest

## 三. 使用

### 1. 导出数据库表结构

* 使用示例：
``` shell
mysql_table_merge -db 'root:password@tcp(127.0.0.1:3306)/?charset=utf8mb4' -mode export -file dbExport.json
```

### 2. 导入数据库表结构

* 使用示例：
``` shell
mysql_table_merge -db 'root:password@tcp(127.0.0.1:3306)/?charset=utf8mb4' -mode import -file dbExport.json
```

### 3. 数据库导入检查

* 在数据库正式导入之前，也可以先使用本命令，检查一下当前的数据库变更，是否允许修改。
* 使用示例：
``` shell
mysql_table_merge -db 'root:password@tcp(127.0.0.1:3306)/?charset=utf8mb4' -mode check -file dbExport.json
```
