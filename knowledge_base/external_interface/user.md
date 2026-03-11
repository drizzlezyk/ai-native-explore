# User External API Documentation

This document describes the external user APIs used by xihe-server for authentication and user management.

## Base URLs

- **Production**: `https://omapi.osinfra.cn`
- **Development**: `https://xiheapi.test.osinfra.cn`

---


## Usage Examples

### cURL Example: Get User Info

```bash
# Variables (replace with actual values)
UT_TOKEN= from cookie "_U_T_"
YG_VALUE= from cookie "_Y_G_"
REFERER="https://mindspore-website.test.osinfra.cn"

curl -X GET "https://xiheapi.test.osinfra.cn/oneid/personal/center/user?community=mindspore" \
  -H "token: ${UT_TOKEN}" \
  -H "cookie: _Y_G_=${YG_VALUE}" \
  -H "Referer: ${REFERER}" \
  -v
```


```json
// response
{
  "data":{
    "company": "xxx",  // 用户所填写的公司名称
    "email": "xxx", // 用户绑定的邮箱账号
    "nickname": "xxx", // 用户填写的昵称
    "phone": "xxx", // 用户绑定的手机号
    "photo": "xxx", // 用户设置的头像
    "signedUp": "xxx", // 用户创建账户的日期
    "username": "xxx", // 用户设置的用户名，大多数时候展示的名字
    "identities": [ // 数组结构，内容为用户绑定的三方账号信息
        {
            "identity": "xxx", // 三方源名称 gitee/github/openatom等
            "accessToken": "xxx", // 三方源认证token
            "login_name": "xxx", // 三方源登录账号名称
            "user_name": "xxx", // 三方源用户名称
        }
    ]
  }
}
```

### cURL Example: Get Manager Token



```go
// infrastructure\authingimpl\email.go
	b := managerBody{
		AppId:     impl.cfg.APPId,
		AppSecret: impl.cfg.Secret,
		GrantType: "token",
	}
```

```bash
APP_ID="your_app_id"
APP_SECRET="your_app_secret"

curl -X POST "https://xiheapi.test.osinfra.cn/oneid/manager/token" \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "token",
    "app_id": "'${APP_ID}'",
    "app_secret": "'${APP_SECRET}'"
  }'
```

```json
// response

{
  "msg":"OK",
  "refresh_token":"7C**3A",
  "token_expire":600,
  "refresh_token_expire":1200,
  "status":200,
  "token":"4075EEF**E"
}
```


### cURL Example: Get User by Manager Token

```bash
UT_TOKEN= from cookie "_U_T_"
YG_VALUE= from cookie "_Y_G_"
MANAGER_TOKEN="<manager_token>"
REFERER="https://mindspore-website.test.osinfra.cn"

curl -X GET "https://xiheapi.test.osinfra.cn/oneid/manager/personal/center/user?community=mindspore" \
  -H "token: ${MANAGER_TOKEN}" \
  -H "user-token: ${UT_TOKEN}" \
  -H "cookie: _Y_G_=${YG_VALUE}" \
  -H "Referer: ${REFERER}" \
  -v
```

```json
// response
{
  "msg":"success",
  "code":200,
  "data":{
    "signedUp":"2026-03-03T08:07:19.933Z",
    "identities":[],
    "phoneCountryCode":"+86",
    "phone":"17796624214",
    "nickname":"",
    "photo":"https://files.authing.co/authing-console/default-user-avatar.png","company":"",
    "privacySignedTime":"2026-03-03 17:46:14",
    "userId":"69a696b7b76ec69b143ed95e",
    "email":"",
    "username":"chenqi479"
    }
}
