# Distributed Hash Table - PPCA 2022
# 简介

本项目使用了chord和kademlia协议实现了分布式哈希表，同时实现了一个由chord协议实现的简易application。

- 参考的资料
  1. [DHT综述](https://luyuhuang.tech/2020/03/06/dht-and-p2p.html)
  2. [chord论文](https://pdos.csail.mit.edu/papers/chord:sigcomm01/chord_sigcomm.pdf)
  3. [kademlia论文](https://pdos.csail.mit.edu/~petar/papers/maymounkov-kademlia-lncs.pdf)
  4. [赵一龙学长的仓库](https://github.com/happierpig/DHT)

## chord 协议
- 封装了4个文件
  1. Node 代表chord协议里的节点
  2. function 实现了一些工具函数
  3. WrapNode 包装了Node类使得符合register规则
  4. network 封装rpc相关，包括服务器监听和远程调用方法
## kademlia 协议
- 封装了6个文件
  1. Node 代表kademlia协议里的节点
  2. function 实现了一些工具函数
  3. WrapNode 包装了Node类使得符合register规则
  4. network 封装rpc相关，包括服务器监听和远程调用方法
  5. RoutingTable 封装routingtable相关
  6. database 对数据进行了封装
## application
设定自身的IP，加入到chord协议的网络中，可以实现上传和下载的功能
- 上传：将指定路径下的文件上传到DHT网络之中，并在另一指定路径之下生成种子文件，并同时输出磁力链接
- 下载：通过种子文件或者磁力链接将目标文件下载到指定路径之下