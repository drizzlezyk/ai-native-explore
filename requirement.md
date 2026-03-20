```json
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
    "total": 1
}
```
```
| 参数名 | 位置 | 类型 | 必填 | 说明 |
| :----| :----: | :----: | :----: | :---- |
| id | query | struct | 否 | 规格筛选（如 "cpu_001"|
| cards_num | query | String | 否 | 规格筛选 如 1 |
| image | query | String | 否 | 镜像别名筛选 |
| page_num | query | Int | 否 | 页码，默认1 |
| page_size | query | Int | 否 | 每页数量，默认20 |
```