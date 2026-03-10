# jupyer在线编程相关接口

```
controller/cloud.go
- rg.GET("/v1/cloud", ctl.List)
- rg.POST("/v1/cloud/subscribe", ctl.Subscribe)
- rg.GET("/v1/cloud/:cid", ctl.Get)  # Jupyter启动websocket接口
- rg.GET("/v1/cloud/pod/:cid", ctl.GetHttp)
- rg.GET("/v1/cloud/read/:owner", ctl.CanRead) # 检查用户是否有读取权限，由jupyter nginx调用
- rg.DELETE("/v1/cloud/pod/:id", ctl.ReleasePod) # 释放Jupyter pod
- rg.GET("/v1/ws/cloud/pod/:id", ctl.WsSendReleasedPod) # 发送已释放的Jupyter pod的websocket消息
```

NPU启动多卡功能只对白名单用户开放，user模块的接口用于校验用户是否有多卡启动权限
```
user/controller/user.go
	rg.GET("/v1/user/whitelist/:type", ctl.CheckWhiteList)
	rg.GET("/v1/user/whitelist", ctl.ListWhitelist)
```

## 专业名词对照表
| 专业名词 | 中文解释 |
| -------- | -------- |
| Jupyter pod | Jupyter pod是指在Kubernetes集群中运行的Jupyter notebook实例 |
| 算力资源 | id | 
| 状态 | 指Jupyter pod当前的运行状态，例如Running、Failed、Succeeded等 |
| 规格 | 指Jupyter pod的规格，具体是CloudSpec|
| 镜像 | 指Jupyter pod运行所使用的Docker镜像名 |
| 创建时间 | 指Jupyter pod创建的时间 |
| 运行时间 | 指Jupyter pod运行的时间 |