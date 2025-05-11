# UserCenter 微服务 - 双协议支持架构

## 架构概述

本服务采用分层架构设计，同时提供 RESTful HTTP-API（基于Gin框架）与 gRPC（原生实现）双协议接入能力。业务逻辑层通过统一Service组件实现协议无关化处理，保障核心业务逻辑的高度复用性。

service使用了协程池与数据对象池，协程池中为session服务与user提供协程使用。

session为自己搭建的一个小型session实现，内置LRU的GC管理，在服务启动与结束时会将记录存储在redis以实现session持久化。

