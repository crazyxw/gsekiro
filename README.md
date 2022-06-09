# gsekiro

这是一个go语言基于websocket实现的[Sekiro Server](https://github.com/virjar/sekiro)


## Quick start
```bash
$ go run *.go
```


## 客户端
### 1. 浏览器接入
浏览器打开 http://127.0.0.1:5612/jsDemo  
你将会在控制台看到client连接信息  

### [2.Java项目接入](https://github.com/crazyxw/SekiroWebSocketDemo)
```java
public class QuickStart {

    public static void main(String[] args) throws URISyntaxException {
        SekiroWebSocketClient webSocketClient = new SekiroWebSocketClient(new URI("ws://127.0.0.1:5612/api/register?group=aaa&clientId=java01&vkey=test"));
        webSocketClient.registerSekiroHandler("hello", new RequestHandler() {
            public void handleRequest(SekiroRequest sekiroRequest, SekiroResponse sekiroResponse) {
                JSONObject jsonObject = sekiroRequest.getRequestJSONObject();
                jsonObject.put("hello", "i com from java");
                sekiroResponse.send(jsonObject);
            }
        });
    }
}
```
### [3. android接入](https://github.com/crazyxw/SekiroWebsocketClientApp)

android和java有些类似,只需要将[sekiro-api](https://github.com/crazyxw/SekiroWebSocketDemo/releases/tag/1.1.0) 引入到项目中即可

## 主动调用
打开一个新的浏览器窗口,访问  
http://127.0.0.1:5612/api/invoke?action=clientTime&group=aaa&vkey=test  
就可以调用client上对应action的函数，并返回相应内容。

## 版本说明
[v1版本]()是基于websocket实现的sekiro server。功能以及API和官方[sekiro-business-demo](https://github.com/virjar/sekiro) 基本相同  
当前版本(main)增加了vkey, 客户端接入和调用都需要设置,增强了安全性。

#### api的区别
main | v1
:------- | :-------
/api/register | /business-demo/register
/api/invoke | /business-demo/invoke
/api/clientQueue | /business-demo/clientQueue
/api/groupList | /business-demo/groupList

## 注意事项  
线上环境请设置相对复杂的vkey,以提升安全性
