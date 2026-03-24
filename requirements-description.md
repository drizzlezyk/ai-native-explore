请你完成下列需求的分析、设计、开发，并输出对接交付件到result目录下,需要分析和设计可以使用knowledge_base目录下的文件，请你直接完成需求分析、设计、开发，中间无需咨询我，且请你确保代码能编译通过，中途你可以优化skill
## 背景
本服务的Jupyter在线编程功能需要合并至官网，其余功能下线（下线的功能不需要删除，只需要关闭接口入口即可）

## 功能需求描述
1. 鉴权方式修改：原有基于token的鉴权方式，修改为基于cookie的鉴权方式
2. 接口下线：下线所有除了Jupyter在线编程功能的所有基于token的接口
3. 新增接口：新增给每个用户查询Jupyter pod运行记录的接口

## 部分接口保留并修改鉴权
Jupyter在线编程相关http接口鉴权方式修改为基于cookie的鉴权方式，具体实现为：
   1. 普通http请求使用中间件刷新cookie
   2. 具体刷新方案：取cookie中的_U_T_，_Y_G_后，调用账号系统Get User Info接口刷新
Jupyter在线编程相关websocket接口不基于cookie刷新，具体实现为：
   1. 注意：websocket请求不使用中间件刷新cookie（因为websocket是长连接，中间件会在每次请求都刷新cookie，会导致鉴权失败）
   2. 具体方案：该接口获取用户信息通过Get User by Manager Token接口，这个接口的鉴权需要更高的权限的token，对应token从Get Manager Token接口获取

## 接口下线
1. 下线的接口保证无法被调用，但是保留在代码中，方便后续维护和恢复。

## 新增接口
1. 新增给每个用户查询Jupyter pod运行记录的功能，这个接口应该去查询当前用户的近期1个月的Jupyter pod运行记录，并返回给前端展示，具体字段包含（算力资源、状态、规格、镜像、创建时间、运行时间）。

我希望你帮我开发一个特性，新增特性是：新增给每个用户查询Jupyter pod运行记录的功能，这个接口应该去查询当前用户的近期1个月的Jupyter pod运行记录，并返回给前端展示，具体字段包含（算力资源、状态、规格、镜像、创建时间、运行时间）。你在做需求分析时，可以参考./claude/knowledge_base中的内容。
1. 该接口由后端分页，通过page_num,page_size参数
2. 该接口路由为：/web/v1/cloud/pod/history

```json
// response
{

    "data": [
        {
            "id": "07xxxxxxxxxxxxxa", //pod id
            "cloud_id":"ascend_002",
            "status": "Running",
            "has_holding" : true,
            "access_url":"https://xxxxxxxx.xihe2.test.osinfra.com",
            "spec": {
              "desc":"4u 16G 20G",
              "cards_num":1
            },
            "image": "python:3.9-ms2.5.0",
            "create_time": 1706745600,
            "running_time": "2:30:30"
        }
    ],
    "has_holding":false,//是否有启动中容器
    "page_num": 1,
    "page_size": 20,
    "total": 1,
    "processor": "Ascend-snt9b"
}
```
query
```
| 参数名 | 位置 | 类型 | 必填 | 说明 |
| :----| :----: | :----: | :----: | :---- |
| id | query | struct | 否 | 规格筛选（如 "cpu_001"|
| cards_num | query | String | 否 | 规格筛选 如 1 |
| image | query | String | 否 | 镜像别名筛选 |
| page_num | query | Int | 否 | 页码，默认1 |
| page_size | query | Int | 否 | 每页数量，默认20 |
```

GET /v1/cloud/pod/history
