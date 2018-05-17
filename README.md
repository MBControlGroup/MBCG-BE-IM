# 民兵指挥系统即时通信服务

## WebSocket路径

```
ws://host:port/ws
```

## 通信过程说明

- 客户端请求与服务器建立WebSocket连接，服务端检查连接请求合法性：
    - 服务端判断客户端请求token的合法性（哈希匹配，token是否过期）
    - 服务端检查token中的用户id是否已经与服务端建立连接，若有则丢弃请求（也可以选择重新建立连接，此处可以进一步讨论）
    - 检查未发现问题，建立连接，将连接加入连接管理器
- 连接建立，通信开始：
    - 客户端发出消息
    - 服务端解析消息类型：
        - 如果是聊天消息：
            - 查找接收端（如果发送端与接收端相同则丢弃消息）
            - 接收端在线则发送消息，之后向发送端发送接收端接收成功与否的状态回执
            - 接收端不在线，将报错消息发给发送端
        - 其他消息类型待后续完善
    - 消息解析出错，返回报错信息

## 通信JSON格式说明

### 聊天

#### 发送端发出消息

```javascript
{
    "type": "chat",  //消息类型，string类型，聊天时为"chat"
    "dst_id": 123,  //目标用户的user_id，uint32类型
    "msg_id": 456,  //发出的消息的id，uint32类型
    "data": "abcdefg"  //消息数据，聊天时为string类型
}
```

#### 接收端收到的消息

```javascript
{
    "type": "chat",  //消息类型，string类型，聊天时为"chat"
    "src_id": 123,  //源用户的user_id，uint32类型
    "data": "abcdefg"  //消息数据，聊天时为string类型
}
```

#### 发送端消息发送状态确认消息

```javascript
{
    "type": "chat",  //消息类型，string类型，聊天时为"chat"
    "dst_id": 123,  //目标用户的user_id，uint32类型
    "msg_id": 456,  //发出的消息的id，uint32类型
    "success": true  //发送成功与否，bool类型
}
```

## IM服务设计说明

- 本服务由连接管理器`ConnectionMannager`模块和业务逻辑部分组成
- 设计上连接管理器与业务逻辑充分解耦，业务逻辑需要封装成函数对象，作为参数传给连接管理器实例获取函数`util.GetManagerInstance()`
- 业务逻辑接口示例：
    ```go
    func messageProcessor(srcID uint32, messageType int, data []byte) {
        // business logic
    }
    ```
- 底层实现上面，每将一个连接加入连接管理器，管理器都会为连接对象分别开启一个读消息协程和一个写消息协程：
    - 读协程会循环读取消息，一旦读取成功，则调用上述业务逻辑函数处理读取到的消息（目前的设计暂时为阻塞处理而不是开启新的协程）
    - 业务逻辑中用户可以调用连接管理器的异步写函数`AsyncWrite`和`AsyncWriteNoCallback`，异步写操作函数会将数据通过channel传递给接收端连接的写消息协程，并且在写操作完成或者出错之后根据用户的选择调用回调函数
    - 上述的聊天消息处理上，回调函数会向消息发送端发送一个回执消息，报告发送消息操作是否成功
