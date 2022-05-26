package main

import "github.com/gorilla/websocket"

type Client struct {
	conn       *websocket.Conn
	channelMap map[string]chan []byte
}
