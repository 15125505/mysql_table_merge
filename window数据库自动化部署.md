# windows下mariadb自动化部署说明

## 一、准备工作

1. 准备好git bash；
2. 为git bash添加wget命令（前往wget官网下载exe文件放到git bash的bin目录下）；

## 二、创建对应的脚本

*  使用git bash进入打算存放数据库的目录；

*  执行以下脚本，创建param.sh

``` shell
echo '# 需要配置的参数
rootPassword=thisIsALongEnoughPassword
defaultUser=green
defaultUserPassword=token2User

# 获取脚本所在的路径
scriptDir=$(pwd)

# 定义一些环境变量
VERSION=10.6.7
FILE=mariadb-$VERSION-winx64

# 添加PATH信息
export PATH=$PATH:$scriptDir/mariadb/$FILE/bin:$scriptDir/mariadb/$FILE/scripts' > param.sh
```

* 执行以下脚本，创建init.sh

``` shell
echo 'cd $(dirname $0)
source ./param.sh

# 创建目录
mkdir $scriptDir/mariadb
if [ $? -ne 0 ]; then
    echo "需要创建的数据库目录已经存在！"
    exit 1
fi
echo "--数据库目录创建成功！"


# 下载并解压，之后删除下载包
wget https://mirrors.xtom.jp/mariadb/mariadb-$VERSION/winx64-packages/$FILE.zip
if [ $? -ne 0 ]; then
    echo "下载失败"
    exit 1
fi
echo "--文件下载成功！"

# 解压
unzip $FILE.zip -d $scriptDir/mariadb
if [ $? -ne 0 ]; then
    echo "解压失败"
    exit 1
fi
rm $FILE.zip
echo "--解压完成！"

# 执行初始化（设置root用户的密码，创建的文件在当前目录的data文件夹之下）
mysql_install_db --datadir="$scriptDir/mariadb/data" --password=$rootPassword
if [ $? -ne 0 ]; then
    echo "初始化失败"
    exit 1
fi
echo "--数据库安装完成！"


# 后台启动服务
mysqld  --defaults-file="$scriptDir/mariadb/data/my.ini" &
echo "--数据库已经在后台启动..."
sleep 3s

# 创建自己的用户（为了和linux上保持一致，因为linux上不是用root用户创建）' > init.sh
echo "mysql -uroot -p\$rootPassword -e\"CREATE USER '\$defaultUser'@'%' IDENTIFIED BY '\$defaultUserPassword';GRANT ALL ON *.* TO '\$defaultUser'@'%'\"" >> init.sh
echo 'if [ $? -ne 0 ]; then
    echo "创建用户失败"
    exit 1
fi
echo "--创建用户成功！"

# 创建完成后，停止服务
mysqladmin -u$defaultUser -p$defaultUserPassword shutdown
if [ $? -ne 0 ]; then
    echo "停止服务失败"
    exit 1
fi

echo "--初始化成功完成！"' >> init.sh

```

* 执行以下脚本，创建start.sh


``` shell
echo 'cd $(dirname $0)
source ./param.sh

# 开始服务
mysqld  --defaults-file="$scriptDir/mariadb/data/my.ini" &
echo "启动服务完成"
' > start.sh
```

* 执行以下脚本，创建stop.sh

``` shell
echo 'cd $(dirname $0)
source ./param.sh

# 停止服务
mysqladmin -u$defaultUser -p$defaultUserPassword shutdown
echo "停止服务完成"' > stop.sh
```

* 为脚本加入执行权限

``` shell
chmod +x param.sh
chmod +x init.sh
chmod +x start.sh
chmod +x stop.sh
```

## 三、数据库的使用

1. 执行start.sh，可以启动数据库服务；
2. 执行stop.sh，可以停止数据库服务；