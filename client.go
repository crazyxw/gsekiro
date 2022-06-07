package main

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Client struct {
	conn       *websocket.Conn
	channelMap map[string]chan []byte
	rwLock     sync.RWMutex
}
