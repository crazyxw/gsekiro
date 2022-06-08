package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
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
	defaultInvokeTimeout = config.Web.InvokeTimeout
	vKey = config.Web.VKey
	lumberLogger := &lumberjack.Logger{
		Filename:  config.Log.Filename,
		MaxAge:    config.Log.MaxAge,
		Compress:  true,
		LocalTime: true,
	}
	multiWriter := io.MultiWriter(os.Stdout, lumberLogger)
	log.SetOutput(multiWriter)
}

var UP = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	//CheckOrigin: func(r *http.Request) bool {
	//	return true
	//},
}

type SekiroRequest struct {
	ReqId string `json:"__sekiro_seq__"`
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

	if !group.addClient(conn, clientId) {
		conn.Close()
		log.Println("设备id已注册, group:" + groupId + " clientId:" + clientId)
		return
	}
	log.Println("注册成功: group:" + groupId + " clientId:" + clientId)
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			break
		}
		log.Println("group:"+groupId+" clientId:"+clientId+" recv:", string(p))
		var sreq SekiroRequest
		parseErr := json.Unmarshal(p, &sreq)
		if parseErr != nil {
			continue
		}
		if sreq.ReqId != "" {
			if reqChan, ok := group.clientMap[clientId].channelMap.LoadAndDelete(sreq.ReqId); ok {
				chain, ok := reqChan.(chan []byte)
				if ok {
					chain <- p
				}
			} else {
				log.Println("对话已结束" + sreq.ReqId)
			}
		}
	}
	group.removeClient(clientId)
	log.Println("服务关闭")
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

func invoke(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	action := params.Get("action")
	groupId := params.Get("group")
	clientId := params.Get("clientId")
	invokeTimeoutStr := params.Get("invoke_timeout")
	var invokeTimeout int = defaultInvokeTimeout
	if invokeTimeoutStr != "" {
		ret, err := strconv.Atoi(invokeTimeoutStr)
		if err == nil {
			invokeTimeout = ret
		}
	}
	group, ok := groupMap[groupId]
	rMap := map[string]string{}
	if ok {
		if len(group.clients) == 0 {
			rMap["msg"] = "no device online"
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
					reqMap := map[string]string{}
					reqMap["__sekiro_seq__"] = req_id
					parseValues(reqMap, r.URL.Query())
					parseValues(reqMap, r.Form)
					cl.rwLock.Lock()
					cl.conn.WriteMessage(websocket.TextMessage, []byte(MapToJson(reqMap)))
					cl.rwLock.Unlock()

					select {
					case p := <-req_chan: // 收到消息返回给客户端
						w.Write(p)
						return
					case <-time.After(time.Second * time.Duration(invokeTimeout)):
						rMap["msg"] = "调用超时"
					}
				} else {
					rMap["msg"] = "请检查action是否正确"
				}
			} else {
				rMap["msg"] = "没有这个client"
			}
		}

	} else {
		rMap["msg"] = "请检查groupId是否正确"
	}
	w.Write([]byte(MapToJson(rMap)))
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
	mux.HandleFunc("/business-demo/register", register)
	mux.HandleFunc("/business-demo/invoke", invoke)
	mux.HandleFunc("/business-demo/clientQueue", getClients)
	mux.HandleFunc("/business-demo/groupList", getGroups)

	server := &http.Server{
		Addr:    config.Web.Port,
		Handler: mux,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
