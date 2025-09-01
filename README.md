# Redis Cluster Manager

A cluster manager for redis cluster, to display cluster status && execute cmd on a set of it's members.

[Readme.md: 简体中文](README_zh.md)

## Introduction
Name: `rcm`, the first arg is a seed node(`IP:PORT`):
- cluster status
```
rcm cluster status 127.0.0.1:6379 -a "password"
# use `-s` or `--show-slots` to show slot ranges
rcm cluster status 127.0.0.1:6379 -a "password" -s
```
The output was grouped by shard，master/slave in a shard will be displayed together, all the shard was ordered by it's
master's addr.

Output:
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
- cluster exec
```
# exec on seed node only:
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING"
# exec on provided nodes:
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -n "1.1.1.1:6379,1.1.1.2:6379"
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -n "90c7...,8f25..."  # can be node ID
# exec on all master nodes/slave nodes:
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -r master
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -r slave
# exec on all nodes:
rcm cluster exec 127.0.0.1:6379 -a "password" -c "PING" -r all
```
Options -n and -c are mutually exclusive, Run `rcm -h` for usage information.

output of cluster exec:
```text
Output of `ping` on 1.1.1.1:6379 :
PONG
Output of `ping` on 1.1.1.1:6380 :
PONG
Done!
```

## Installation
Linux:

There are two ways to install `rcm`:
- Build from source:
```
git clone https://code.wifi.com/mysql/redis-cluster-manager.git
cd redis-cluster-manager
make
```
After `make`, an executable file named `rcm` will be generated in `/usr/local/bin/` by default.
- Download rcm.zip from Releases  

Then unzip it to `/usr/local/bin/` or any other directory in your PATH.

## TODO：
#### 1) Improvements
- [x] display keysCount for cluster status
- [x] cluster exec: when use -n option, display `(node ID)` after addr
- [x] cluster exec: if both -n and -r are not given, exec on seed node only, add new role `all` to exec on all nodes
- [x] display error nodes in the end
- [x] add a (init) flag for slaves with master_sync_in_progress=1
- [x] check if slots count=16384 for cluster status
#### 2) New Features
- [x] add support for master-slave cluster
- [ ] add cluster slowlog parser
- [ ] add cluster keymap, displays histogram distributions of keys across different length ranges. You can specify a 
comma-separated list of bucket boundaries and a sampling rate for keymap.

## FAQ