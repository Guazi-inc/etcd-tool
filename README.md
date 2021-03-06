# etcd-tool

* 从json配置文件转化为etcd的key,value对
* 复制etcd的数据
* 从etcd读取配置


## 配置读取

### key-value结构
    对于下面 redis 的配置，转化为两对etcd中key-value
```json
{
    "redis": {
        "address": "localhost:6379",
        "prefix": "test"
    }
}

"/redis/address": "localhost:2379"
"/redis/prefix": "test"
```


### SDK使用
#### 设置etcd的连接地址: username:password@addr1,addr2/namespace
```
example: root:rootpw@localhost:2379,localhost2479/my_group/my_project
```

方案一：环境变量
```shell
ETCD_ADDR=localhost:2379
```
方案二：调用函数 
```go
import "github.com/Guazi-inc/etcd-tool/config"

config.InitETCD("localhost:2379")
```

#### Get Config
```go
import "github.com/Guazi-inc/etcd-tool/config"

var cfg struct{
    Address string `json:"address"`
    Prefix  string `json:"prefix"`
}

//在根目录下读取配置，full_key: /redis
err := config.Get("/redis", &cfg)
```

##### Get Config In Namespace
```
//在一级namespace下读取配置，full_key: /my_group/redis
err := config.GetInNamespace("/redis", &cfg, 1)

//在二级namespace下读取配置，full_key: /my_group/my_project/redis
err = config.GetInNamespace("/redis", &cfg, 2)

//更多级的namespace...
```

#### Watch机制
    config.Get方法本身已经使用watch机制对数据做了缓存，为了减少反射带来的开销，以及定制化的需求提供自定义的watch入口，可以在某些key变化时执行自定义的方法

    下面的方法表示在 "/redis" 前缀的key发生变化时，执行一个function
```go
import "github.com/Guazi-inc/etcd-tool/config"

config.WithCustomWatch("/redis", func() {})
```


#### 配置检查
    在程序启动前通过调用下面方法可以进行配置检查，若不满足条件则会panic

```go
import "github.com/Guazi-inc/etcd-tool/config"

//若存在值为空的key，则会panic
keys := []string{"/test/key1"}

//若以key为前缀没有数据，则会panic
keysWithPrefix := []string{"/test/dir"}

config.CheckKeys(keys, keysWithPrefix)

```