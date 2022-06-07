package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const defaultInvokeTimeout = 3

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
			//if reqChan, ok := group.clientMap[clientId].channelMap[sreq.ReqId]; ok {
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
					//cl.channelMap[req_id] = req_chan
					cl.channelMap.Store(req_id, req_chan)
					reqMap := map[string]string{}
					reqMap["__sekiro_seq__"] = req_id
					parseValues(reqMap, r.URL.Query())
					parseValues(reqMap, r.Form)
					cl.conn.WriteMessage(websocket.TextMessage, []byte(MapToJson(reqMap)))
					//defer cl.channelMap.Delete(req_id)

					select {
					case p := <-req_chan: // 收到消息返回给客户端
						//fmt.Println("write:" + string(p))
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
	http.HandleFunc("/", index)
	http.HandleFunc("/jsDemo", jsDemo)
	http.HandleFunc("/business-demo/register", register)
	http.HandleFunc("/business-demo/invoke", invoke)
	http.HandleFunc("/business-demo/clientQueue", getClients)
	http.HandleFunc("/business-demo/groupList", getGroups)
	err := http.ListenAndServe(":5612", nil)
	if err != nil {
		panic(err)
	}
}
