# Redis Cluster Manager

A cluster manager for Redis Cluster and Redis master-slave. It displays topology/status information and executes Redis commands on selected members. Some high-risk Redis commands are forbidden.

[Readme.md: 简体中文](README_zh.md)

## Introduction
Name: `rcm`. Cluster subcommands take a seed node in `IP:PORT` format.
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
NodeID                                       Address                 Role            Memory(GB)      KeysCount       Clients         Slots       SlotRanges
------                                       -------                 ----            ----------      ---------       -------         -----       ----------
90c7c50bf195ba10e2fbf5a90d12b2ed570e3352     1.1.1.1:6379            master          0.24/10.00      1024            29/20000        5461        ...
57c63639108496dd5349863a9589408a7f5b385c     1.1.1.2:6379            -slave          0.24/10.00      1024            12/20000
ef2ad9890ab216c311de4f66995bbcb72bada047     1.1.1.1:6380            master          0.24/10.00      2048            31/20000        5462        ...
8f259674d2742cbcbdaf23c070e032c368090c83     1.1.1.2:6380            -slave(init)    0.24/10.00      2048            15/20000
...         
Total up masters in cluster: 3
Total up members in cluster: 6
```
- cluster exec
```
# exec on seed node only:
rcm cluster exec 127.0.0.1:6379 -a "password" -- PING
# exec on provided nodes:
rcm cluster exec 127.0.0.1:6379 -a "password" -n "1.1.1.1:6379,1.1.1.2:6379" -- PING
rcm cluster exec 127.0.0.1:6379 -a "password" -n "90c7...,8f25..." -- PING  # can be node ID
# exec on all master nodes/slave nodes:
rcm cluster exec 127.0.0.1:6379 -a "password" -r master -- PING
rcm cluster exec 127.0.0.1:6379 -a "password" -r slave -- PING
# exec on all nodes:
rcm cluster exec 127.0.0.1:6379 -a "password" -r all -- PING

# commands with arguments:
rcm cluster exec 127.0.0.1:6379 -a "password" -- GET mykey
rcm cluster exec 127.0.0.1:6379 -a "password" -- SET mykey myvalue
```

The Redis command is positional; there is no `-c` option. Use `--` before the Redis command when the command or its arguments come after `rcm` flags. Options `-n` and `-r` are mutually exclusive. Run `rcm cluster exec -h` for usage information.

Forbidden commands: `DEBUG`, `FLUSHALL`, `FLUSHDB`, `SHUTDOWN`, `MONITOR`.

output of cluster exec:
```text
Output of `[PING]` on 1.1.1.1:6379:
PONG
Output of `[PING]` on 1.1.1.1:6380:
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
- [ ] add cluster slowlog parser, collect slowlogs from all nodes and display in a unified way.
- [ ] add instance monitor parser, collect `monitor` result from seed node and display cmd distribution
- [ ] add instance keymap, displays histogram distributions of keys across different length ranges. You can specify a 
comma-separated list of bucket boundaries and a sampling rate for keymap.

## FAQ
