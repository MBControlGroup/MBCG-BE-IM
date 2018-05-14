# 民兵指挥系统即时通信服务


## 通信JSON格式说明

### 聊天

#### 发送端发出消息

```json
{
    "type": "chat",  //消息类型，string类型，聊天时为"chat"
    "dst_id": 123,  //目标用户的user_id，uint32类型
    "msg_id": 456,  //发出的消息的id，uint32类型
    "data": "abcdefg"  //消息数据，聊天时为string类型
}
```

#### 接收端收到的消息

```json
{
    "type": "chat",  //消息类型，string类型，聊天时为"chat"
    "src_id": 123,  //源用户的user_id，uint32类型
    "data": "abcdefg"  //消息数据，聊天时为string类型
}
```

#### 发送端消息发送状态确认消息

```json
{
    "type": "chat",  //消息类型，string类型，聊天时为"chat"
    "dst_id": 123,  //目标用户的user_id，uint32类型
    "msg_id": 456,  //发出的消息的id，uint32类型
    "success": true  //发送成功与否，bool类型
}
```