# jupyer在线编程相关接口

```
controller/cloud.go
- rg.GET("/v1/cloud", ctl.List)
- rg.POST("/v1/cloud/subscribe", ctl.Subscribe)
- rg.GET("/v1/cloud/:cid", ctl.Get)
- rg.GET("/v1/cloud/pod/:cid", ctl.GetHttp)
- rg.GET("/v1/cloud/read/:owner", ctl.CanRead)
- rg.DELETE("/v1/cloud/pod/:id", ctl.ReleasePod)
- rg.GET("/v1/ws/cloud/pod/:id", ctl.WsSendReleasedPod)
```


```
user/controller/user.go
	rg.GET("/v1/user/whitelist/:type", ctl.CheckWhiteList)
	rg.GET("/v1/user/whitelist", ctl.ListWhitelist)
```