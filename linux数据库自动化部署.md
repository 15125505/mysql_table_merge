# linux下mariadb自动化部署说明

## 一、创建数据库用户

* 切换到root用户
* 创建数据库专属用户
``` shell
useradd green
passwd green
# 输入密码
```
* 为用户设置sudo权限
    * 输入visudo，找到root ALL=(ALL) ALL
    * 在下面依葫芦画瓢添加一行green ALL=(ALL) ALL
    * :wq保存并退出

## 二、准备对应的脚本并完成数据库安装


``` shell

# 切换用户
su green

# 安装必须要用的软件
sudo yum install wget  tar vim -y

# 创建存放数据库的目录（如果需要使用其它磁盘，那么前往该磁盘创建该目录）
mkdir /home/green/myDb 
cd /home/green/myDb

# 创建param脚本(如果需要修改用户登录数据库的密码，可以在该脚本中修改)
echo '# 需要配置的参数
defaultUserPassword=yourPassword

# 获取脚本所在的路径
scriptDir=$(pwd)

# 定义一些环境变量
VERSION=10.6.7
FILE=mariadb-$VERSION-linux-systemd-x86_64

# 添加PATH信息
export PATH=$PATH:$scriptDir/mariadb/$FILE/bin:$scriptDir/mariadb/$FILE/scripts' > param.sh

#  创建init.sh
echo 'cd $(dirname $0)
source ./param.sh

# 创建目录
mkdir $scriptDir/mariadb
if [ $? -ne 0 ]; then
    echo "需要创建的数据库目录已经存在！"
    exit 1
fi

# 下载并解压，之后删除下载包
wget https://mirrors.xtom.jp/mariadb/mariadb-$VERSION/bintar-linux-systemd-x86_64/$FILE.tar.gz
if [ $? -ne 0 ]; then
    echo "下载失败"
    exit 1
fi

# 解压
tar -zxvf $FILE.tar.gz -C $scriptDir/mariadb
if [ $? -ne 0 ]; then
    echo "解压失败"
    exit 1
fi
rm $FILE.tar.gz

# 执行初始化（默认使用当前用户，创建的文件在当前目录的data文件夹之下）
mysql_install_db --basedir=$scriptDir/mariadb/$FILE --datadir=$scriptDir/mariadb/data

# 后台启动服务
mysqld_safe --no-defaults --basedir=$scriptDir/mariadb/$FILE --datadir=$scriptDir/mariadb/data &
sleep 3s

# 设置密码（设置密码之后，非当前用户才能访问数据库）
mysqladmin password $defaultUserPassword
if [ $? -ne 0 ]; then
    echo "设置用户密码失败"
    exit 1
fi

# 停止服务
mysqladmin shutdown

echo "初始化完成"
' > init.sh

# 创建start.sh
echo 'cd $(dirname $0)
source ./param.sh

# 启动服务
mysqld_safe --no-defaults --basedir=$scriptDir/mariadb/$FILE --datadir=$scriptDir/mariadb/data &
sleep 1s
echo "服务已经在后台启动"
' > start.sh


# 创建stop.sh
echo 'cd $(dirname $0)
source ./param.sh

# 停止服务
mysqladmin shutdown
' > stop.sh

# 完成数据库初始化
sh init.sh

```

## 三、将数据库部署为服务，并使其开机自动启动

**注意，如果上面修改过数据目录，那么下面的脚本目录也需要进行相应的修改。**

``` shell
sudo vim /lib/systemd/system/greenDb.service # 创建服务

# 输入以下内容
[Unit]
Description=greenDb
After=network.target
[Service]
Type=forking
User=green
Group=green
ExecStart=/bin/bash /home/green/myDb/start.sh 
ExecReload=
ExecStop=/bin/bash /home/green/myDb/stop.sh
PrivateTmp=false
TimeoutStartSec=600
[Install]
WantedBy=multi-user.target

# 设置权限
sudo chmod 754 /lib/systemd/system/greenDb.service # 设置权限
sudo systemctl enable greenDb # 保证开机启动

# 添加一个mysql快捷命令，方便测试
sudo ln -s /home/green/myDb/mariadb/mariadb-10.6.7-linux-systemd-x86_64/bin/mysql /usr/bin/mysql

# 重启电脑
sudo reboot

# 重启之后，验证
sudo mysql # 如果能正常进入mysql命令行，说明一切OK
```

## 四、使用说明

* 查看服务状态： `sudo systemctl status greenDb`
* 停止服务： `sudo systemctl stop greenDb`
* 开始服务： `sudo systemctl start greenDb`
* 进入数据库：`sudo mysql`


