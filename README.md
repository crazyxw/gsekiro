# gsekiro

这是一个go语言基于websocket实现的[Sekiro Server](https://github.com/virjar/sekiro)


## Quick start
```bash
$ go run *.go
```
接下来浏览器打开 http://127.0.0.1:5612/jsDemo  
你将会在控制台看到client连接信息  
另外打开一个新的浏览器窗口,访问
 http://127.0.0.1:5612/api/invoke?action=clientTime&group=aaa&vkey=test  
就可以调用client上对应action的函数，并返回相应内容。

# 注意事项  
线上环境请设置相对复杂的vkey,以提升安全性
