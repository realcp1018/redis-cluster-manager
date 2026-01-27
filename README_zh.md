# Redis Cluster Manager

一个用 Go 语言编写的 Redis 集群管理工具，支持集群状态展示，并发执行 Redis 指令的功能(禁用了一些高危指令)。

## 基本功能
我们把命令行工具命名为 `rcm`，其首个参数为seed node(格式为`IP:PORT`)：
- 集群状态展示
```
rcm cluster status 127.0.0.1:6379 -a "password"
# 添加`-s`或`--show-slots`展示详细的slot分布信息
rcm cluster status 127.0.0.1:6379 -a "password" -s
```
输出结果按shard分组，master/slave会显示在一起，同时shard展示按master地址进行排序，同一个shard内的slave也是按地址排序。

输出示例：
```text
=======================================================================================================
Cluster Version:    7.0.9
=======================================================================================================
NodeID                                       Addr                    Role      Mem(GB)         Client          
------                                       ----                    ----      -------         ------          
90c7c50bf195ba10e2fbf5a90d12b2ed570e3352     1.1.1.1:6379            master    0.24/10.00      29/20000        
57c63639108496dd5349863a9589408a7f5b385c     1.1.1.2:6379            -slave    0.24/10.00      12/20000                    
ef2ad9890ab216c311de4f66995bbcb72bada047     1.1.1.1:6380            master    0.24/10.00      31/20000       
8f259674d2742cbcbdaf23c070e032c368090c83     1.1.1.2:6380            -slave    0.24/10.00      15/20000                    
...         
Total nodes in cluster: 6
Total shard in cluster: 3
```
- 并发执行 Redis 指令
```
# 仅在seed节点执行：
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING"
# 在指定节点执行：
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -n "1.1.1.1:6379,1.1.1.2:6379"
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -n "90c7...,8f25..."  # 节点ID
# 在所有master/slave节点执行：
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -r master
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -r slave
# 在所有节点执行：
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -r all
```
-n与-c参数互斥，使用 `rcm -h 查看帮助。
输出示例:
```text
Output of `ping` on 1.1.1.1:6379 :
PONG
Output of `ping` on 1.1.1.1:6380 :
PONG
Done!
```

## 安装部署
Linux:
```
git clone https://code.wifi.com/mysql/redis-cluster-manager.git
cd redis-cluster-manager
make
```
make后自动生成可执行文件 `rcm`，默认存放在`/usr/local/bin/`目录下。

你也可以手动build:
```
go build -o rcm main.go
```
手动build出来的文件不包含版本信息等。

或者也可以直接从Releases下载rcm.zip，解压后即可使用。

## TODO：
#### 1) 改进
- [x] 支持keysCount展示
- [x] exec 指定-n参数时，地址输出时增加`(node ID)`后缀
- [x] exec 不指定-n和-r参数时，默认在seed节点执行指令，新增-r all表示在所有节点执行指令
- [x] 异常实例信息展示在末尾
- [x] 处于全量同步状态的slave role之后新增(init)标识，表示处于全量同步过程中
- [x] 增加slots总数校验
#### 2) 新功能
- [x] 增加对主从集群的支持
- [ ] 为cluster增加slowlog分析功能
- [ ] 为instance增加新的keymap功能，以直方图形式展示keys在不同长度范围的分布，支持输入逗号分隔的buckets列表，支持采样率设置
- [ ] 增加rcm cluster check命令，在集群所有节点执行cluster nodes指令，结果排序后去重，找出不一致的节点信息