# Redis Cluster Manager

一个用 Go 语言编写的 Redis 集群管理工具，支持集群状态展示，并发执行 Redis 指令的功能(禁用了一些高危指令)。

## 1. 基本功能
我们把命令行工具命名为 `rcm`，其首个参数为seed node(格式为`IP:PORT`)：
- 集群状态展示
```
rcm cluster status 127.0.0.1:6379 -a "password"
# 添加`-s`或`--show-slots`展示详细的slot分布信息
rcm cluster status 127.0.0.1:6379 -a "password" -s
```
- 并发执行 Redis 指令
```
# 在所有节点执行：
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING"
# 在指定节点执行：
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -n "1.1.1.1:6379,1.1.1.2:6379"
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -n "90c7...,8f25..."  # 节点ID
# 在所有master/slave节点执行：
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -r master
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -r slave
```
-n与-c参数互斥，使用 `rcm -h 查看帮助。

## 2. 安装部署
```
git clone https://github.com/realcp1018/redis-cluster-manager.git
cd redis-cluster-manager
make
```
make后自动生成可执行文件 `rcm`，默认存放在`/usr/local/bin/`目录下。

你也可以手动build:
```
go build -o rcm main.go
```
手动build出来的文件不包含版本信息等。

## 3. TODO：
- [ ] exec 指定-n参数时，地址输出时增加`(node ID)`后缀
- [ ] exec 不指定-n和-r参数时，默认在seed节点执行指令，新增-r all表示在所有节点执行指令