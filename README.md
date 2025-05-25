# GoRedis 🚀

一个用 Go 语言实现的 Redis 服务器，支持 Redis 协议和主要数据结构操作。

## 项目简介 📖

GoRedis 是一个高性能的 Redis 兼容服务器实现，使用 Go 语言编写。它实现了 Redis 的核心功能，包括：

- 🔗 **RESP 协议支持** - 兼容 Redis 客户端
- 📦 **多种数据结构** - 字符串、哈希表、列表、集合、有序集合
- 💾 **持久化支持** - AOF (Append Only File) 持久化
- 🗃️ **多数据库支持** - 支持多个独立的数据库实例
- 🌐 **集群功能** - 支持分布式集群部署
- 🚄 **TCP 服务器** - 高性能的网络服务

## 功能特性 ⭐

### 数据结构支持 🗂️
- 📝 **字符串 (Strings)**
- 🗄️ **哈希表 (Hashes)**
- 📃 **列表 (Lists)**
- 🎯 **集合 (Sets)**：支持底层从 intset 自动切换到 hashmap
- 🏆 **有序集合 (Sorted Sets)**：支持底层从 listpack 自动切换到 ziplist + skiplist

### 核心功能 🔧
- 🔄 **数据库选择** - SELECT 命令支持多数据库
- 📝 **AOF 持久化** - 数据持久化到磁盘
- 🌍 **集群支持** - 分布式部署和数据分片

## 项目结构 📁

```
goredis/
├── main.go              # 程序入口点
├── redis.conf           # 配置文件
├── go.mod              # Go 模块依赖
├── database/           # 数据库核心实现
│   ├── database.go     # 数据库操作
│   ├── command.go      # 命令操作
│   ├── standalone_database.go  # 单机数据库
│   ├── db.go           # 数据库操作
│   ├── strings.go      # 字符串操作
│   ├── hash.go         # 哈希表操作
│   ├── lists.go        # 列表操作
│   ├── set.go          # 集合操作
│   ├── zset.go         # 有序集合操作
│   └── keys.go         # 键管理操作
├── RESP/               # Redis 协议实现
│   ├── handler/        # 请求处理器
│   ├── reply/          # 响应格式
│   ├── parser/         # 协议解析
│   ├── connection/     # 连接管理
│   └── client/         # 客户端实现
├── datastruct/         # 数据结构实现
│   ├── dict/           # 字典实现
│   ├── skiplist/       # 跳表实现
│   ├── set/            # 集合实现
│   ├── hash/           # 哈希表实现
│   └── zset/           # 有序集合实现
├── cluster/            # 集群功能
│   ├── cluster_database.go  # 集群数据库
│   ├── router.go       # 路由管理
│   └── client_pool.go  # 客户端连接池
├── TCP/                # TCP 服务器
├── aof/                # AOF 持久化
├── config/             # 配置管理
├── interface/          # 接口定义
└── lib/                # 工具库
```

## 快速开始 🚀

### 环境要求 📋
- Go 1.23.3 或更高版本

### 安装和运行 ⚡

1. **📥 克隆项目**
```bash
git clone <repository-url>
cd goredis
```

2. **📦 安装依赖**
```bash
go mod tidy
```

3. **⚙️ 配置服务器**
编辑 `redis.conf` 文件：
```conf
bind 0.0.0.0        # 绑定地址
port 6380           # 监听端口
databases 16        # 数据库数量
appendonly yes      # 启用 AOF 持久化
appendfilename appendonly.aof  # AOF 文件名
```

4. **🚀 启动服务器**
```bash
go run main.go
```

服务器将在配置的端口上启动（默认 6380）。

### 连接测试 🔗

使用 Redis 客户端连接：
```bash
redis-cli -p 6380
```

或使用任何支持 Redis 协议的客户端库。

## 配置选项 ⚙️

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `bind` | 服务器绑定地址 | 0.0.0.0 |
| `port` | 服务器监听端口 | 6380 |
| `databases` | 数据库数量 | 16 |
| `appendonly` | 是否启用 AOF 持久化 | yes |
| `appendfilename` | AOF 文件名 | appendonly.aof |

## 支持的命令 💻

### 字符串操作 📝
- `GET key` - 获取键值
- `SET key value` - 设置键值
- `SETNX key value` - 仅当键不存在时设置
- `GETSET key value` - 设置新值并返回旧值
- `STRLEN key` - 获取字符串长度

### 哈希表操作 🗄️
- `HSET key field value` - 设置哈希字段值
- `HGET key field` - 获取哈希字段值
- `HEXISTS key field` - 检查哈希字段是否存在
- `HDEL key field [field ...]` - 删除哈希字段
- `HLEN key` - 获取哈希表字段数量
- `HGETALL key` - 获取所有哈希字段和值
- `HKEYS key` - 获取所有哈希字段名
- `HVALS key` - 获取所有哈希字段值
- `HMGET key field [field ...]` - 批量获取哈希字段值
- `HMSET key field value [field value ...]` - 批量设置哈希字段值
- `HSETNX key field value` - 仅当字段不存在时设置
- `HENCODING key` - 获取哈希表编码类型

### 列表操作 📃
- `LPUSH key value [value ...]` - 左侧插入元素
- `RPUSH key value [value ...]` - 右侧插入元素
- `LPOP key` - 左侧弹出元素
- `RPOP key` - 右侧弹出元素
- `LRANGE key start stop` - 获取指定范围的元素
- `LLEN key` - 获取列表长度
- `LINDEX key index` - 获取指定索引的元素
- `LSET key index value` - 设置指定索引的元素值

### 集合操作 🎯
- `SADD key member [member ...]` - 添加成员
- `SCARD key` - 获取集合成员数量
- `SISMEMBER key member` - 检查成员是否存在
- `SMEMBERS key` - 获取所有成员
- `SREM key member [member ...]` - 删除成员
- `SPOP key [count]` - 随机弹出成员
- `SRANDMEMBER key [count]` - 随机获取成员
- `SUNION key [key ...]` - 并集运算
- `SUNIONSTORE destination key [key ...]` - 并集运算并存储
- `SINTER key [key ...]` - 交集运算
- `SINTERSTORE destination key [key ...]` - 交集运算并存储
- `SDIFF key [key ...]` - 差集运算
- `SDIFFSTORE destination key [key ...]` - 差集运算并存储

### 有序集合操作 🏆
- `ZADD key score member [score member ...]` - 添加成员
- `ZSCORE key member` - 获取成员分数
- `ZCARD key` - 获取有序集合成员数量
- `ZRANGE key start stop [WITHSCORES]` - 按排名范围获取成员
- `ZREM key member [member ...]` - 删除成员
- `ZCOUNT key min max` - 统计分数范围内的成员数量
- `ZRANK key member` - 获取成员排名
- `ZTYPE key` - 获取有序集合类型

### 键管理 🗝️
- `PING` - 测试连接
- `DEL key [key ...]` - 删除键
- `SELECT db` - 选择数据库

## 开发指南 👨‍💻

### 添加新命令 ➕

1. 在相应的数据类型文件中实现命令逻辑
2. 在 `database/command.go` 中注册命令
3. 添加相应的测试用例

### 扩展数据结构 🔧

1. 在 `datastruct/` 目录下实现新的数据结构
2. 在 `database/` 目录下添加操作接口
3. 更新命令处理器