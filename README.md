[English Document](./README.en.md)

# sockfwd

一个在socket之间转发数据的小工具。

比如，能够将unix socket通过tcp端口暴露出来，也可以将监听本地的tcp端口通过0.0.0.0暴露出来。

## 用法

```
Usage:
  sockfwd [flags]

Flags:
  -d, --destination string   目的地址，即要转发到的地址
  -s, --source string        源地址，即接收请求的地址
  -q, --quiet                静默模式
```

## 例子

将本地的docker实例暴露到网络上：`./sockfwd -s tcp://127.0.0.1:8090 -d unix:///var/run/docker.sock`

将`127.0.0.1:8080`端口暴露到`0.0.0.0:8090`端口上：`./sockfwd -s tcp://127.0.0.1:8090 -d unix://127.0.0.1:8090`
