package main

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Client struct {
	conn *websocket.Conn
	channelMap sync.Map
	rwLock     sync.RWMutex
}
