package main

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Group struct {
	clients   []string
	clientMap map[string]*Client
}

func (group *Group) removeClient(clientId string) {
	for i := 0; i < len(group.clients); i++ {
		if group.clients[i] == clientId {
			group.clients = append(group.clients[:i], group.clients[i+1:]...)
			group.clientMap[clientId].conn.Close()
			delete(group.clientMap, clientId)
			break
		}
	}
}

func (group *Group) addClient(conn *websocket.Conn, clientId string) bool {
	if group.clientMap == nil {
		group.clientMap = map[string]*Client{}
	}
	if _, ok := group.clientMap[clientId]; ok {
		return false
	} else {
		group.clients = append(group.clients, clientId)
		group.clientMap[clientId] = &Client{
			conn: conn,
			//channelMap: map[string]chan []byte{},
			channelMap: sync.Map{},
		}
		return true
	}
}

// 顺序循环
func (group *Group) getClient() string {
	switch len(group.clients) {
	case 0:
		return ""
	case 1:
		return group.clients[0]
	default:
		client := group.clients[0]
		group.clients = group.clients[1:]
		group.clients = append(group.clients, client)
		return client
	}
}
