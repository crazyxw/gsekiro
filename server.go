package main

import (
	"github.com/gorilla/websocket"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var defaultInvokeTimeout int
var vKey string
var config Config

func init() {
	err := config.loadFromFile()
	if err != nil {
		panic(err)
	}
	defaultInvokeTimeout = int(math.Max(float64(config.Web.InvokeTimeout), 3))
	vKey = config.Web.VKey
	if config.Log.Filename != "" {
		lumberLogger := &lumberjack.Logger{
			Filename:  config.Log.Filename,
			MaxAge:    config.Log.MaxAge,
			Compress:  true,
			LocalTime: true,
		}
		multiWriter := io.MultiWriter(os.Stdout, lumberLogger)
		log.SetOutput(multiWriter)
	}
}

var UP = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	//CheckOrigin: func(r *http.Request) bool {
	//	return true
	//},
}

type Msg struct {
	Type  string
	ReqId string
	Body  []byte
}

func (this *Msg) toBytes() []byte {
	head := []byte(this.Type + this.ReqId)
	return append(head, this.Body...)
}

var groupMap = map[string]*Group{}

func register(w http.ResponseWriter, r *http.Request) {
	reqParams := r.URL.Query()
	var clientId = reqParams.Get("clientId")
	var groupId = reqParams.Get("group")
	if clientId == "" || groupId == "" {
		return
	}
	conn, err := UP.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	group, ok := groupMap[groupId]
	if !ok {
		group = &Group{}
		groupMap[groupId] = group
	}
	msg := Msg{
		Type: "0",
	}
	if !group.addClient(conn, clientId) {
		resp := Response{Code: 1, Msg: "设备id已注册"}
		msg.Body = resp.toJson()
		conn.WriteMessage(websocket.TextMessage, msg.toBytes())
		conn.Close()
		log.Printf("设备id已注册, group:group:%s clientId: %s \n", groupId, clientId)
		return
	}
	resp := Response{Code: 0, Msg: "注册成功"}
	msg.Body = resp.toJson()
	conn.WriteMessage(websocket.TextMessage, msg.toBytes())
	log.Printf("注册成功: group:%s clientId: %s \n", groupId, clientId)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("group:%s clientId %s recv:%s \n", groupId, clientId, string(message))
		reqId := string(message[:36])
		if reqId != "" {
			if reqChan, ok := group.clientMap[clientId].channelMap.LoadAndDelete(reqId); ok {
				chain, ok := reqChan.(chan []byte)
				if ok {
					chain <- message[36:]
				}
			} else {
				log.Println("没有找到此reqId:" + reqId)
			}
		}
	}
	group.removeClient(clientId)
	log.Printf("服务已关闭: group:%s, clientId:%s\n", groupId, clientId)
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, This is a websocket rpc service!"))
}

func getGroups(w http.ResponseWriter, r *http.Request) {
	var result []string
	for group := range groupMap {
		result = append(result, group)
	}
	if len(result) == 0 {
		w.Write([]byte("当前没有任何group"))
	} else {
		w.Write([]byte(strings.Join(result, ", ")))
	}

}

func getClients(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	groupId := params.Get("group")
	if group, ok := groupMap[groupId]; ok {
		if len(group.clients) > 0 {
			w.Write([]byte(strings.Join(group.clients, ", ")))
		} else {
			w.Write([]byte("当前group没有client在线"))
		}
	} else {
		w.Write([]byte("当前group不存在"))
	}
}

type Params map[string]string

func (p Params) getAndDelete(key string) string {
	if value, ok := p[key]; ok {
		delete(p, key)
		return value
	}
	return ""
}

func (p Params) get(key string) string {
	if value, ok := p[key]; ok {
		return value
	}
	return ""
}

func getRequestParams(r *http.Request) Params {
	reqParams := Params{}
	parseValues(reqParams, r.URL.Query())
	parseValues(reqParams, r.Form)
	parseValues(reqParams, r.PostForm)
	return reqParams
}

func invoke(w http.ResponseWriter, r *http.Request) {
	params := getRequestParams(r)
	invokeTimeoutStr := params.getAndDelete("invoke_timeout")
	groupId := params.get("group")
	clientId := params.get(params["client"])
	action := params.get("action")
	delete(params, "vkey")
	var invokeTimeout int = defaultInvokeTimeout
	if invokeTimeoutStr != "" {
		ret, err := strconv.Atoi(invokeTimeoutStr)
		if err == nil {
			invokeTimeout = ret
		}
	}
	group, ok := groupMap[groupId]
	res := Response{Code: 1}
	if ok {
		if len(group.clients) == 0 {
			res.Msg = "no device online"
		} else {
			if clientId == "" {
				clientId = group.getClient()
			}
			cl, ok := group.clientMap[clientId]
			if ok {
				if action != "" {
					req_id := getUuid()
					req_chan := make(chan []byte, 1)
					cl.channelMap.Store(req_id, req_chan)
					cl.rwLock.Lock()
					_msg := Msg{"1", req_id, []byte(MapToJson(params))}
					cl.conn.WriteMessage(websocket.TextMessage, _msg.toBytes())
					cl.rwLock.Unlock()

					select {
					case msg := <-req_chan: // 收到消息返回给客户端
						w.Write(msg)
						return
					case <-time.After(time.Second * time.Duration(invokeTimeout)):
						cl.channelMap.Delete(req_id)
						res.Msg = "调用超时"
					}
				} else {
					res.Msg = "请检查action是否正确"
				}
			} else {
				res.Msg = "没有这个client"
			}
		}

	} else {
		res.Msg = "请检查groupId是否正确"
	}
	w.Write(res.toJson())
}

func jsDemo(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "example/js/demo.html")
}

func main() {
	middlewares := []Middleware{
		VerifySignMiddleware,
		//MetricMiddleware,
	}

	mux := NewMyMux()
	mux.HandleFunc("/", index)
	mux.HandleFunc("/jsDemo", jsDemo)
	mux.Use(middlewares...) // 下面的接口都经过中间件
	mux.HandleFunc("/api/register", register)
	mux.HandleFunc("/api/invoke", invoke)
	mux.HandleFunc("/api/clientQueue", getClients)
	mux.HandleFunc("/api/groupList", getGroups)

	server := &http.Server{
		Addr:    config.Web.Port,
		Handler: mux,
	}
	log.Printf("gsekiro is running. http://0.0.0.0%s", config.Web.Port)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
