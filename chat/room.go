package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kazuhei/goblueprints/trace"
)

type room struct {
	// forwordは他のクライアントに転送するためのメッセージを保持するチャネルです
	forword chan []byte
	// joinはチャットルームに参加しようとしているクライアントのためのチャネルです
	join chan *client
	// leaveはチャットルームに参加しようとしているクライアントのためのチャネルです
	leave chan *client
	// clientsには在籍している全てのクライアントが保存されます
	clients map[*client]bool
	// tracerはチャットルーム上で行われた操作のログを受け取ります
	tracer trace.Tracer
}

func newRoom() *room {
	return &room{
		forword: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// 参加
			r.clients[client] = true
			r.tracer.Trace("新しいクライアントが参加しました")
		case client := <-r.leave:
			// 退室
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("クライアントが退室しました")
		case msg := <-r.forword:
			for client := range r.clients {
				select {
				case client.send <- msg:
					// メッセージを送信
					r.tracer.Trace("クライアントに送信しました")
				default:
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace("送信に失敗しました。クライアントをクリーンアップします")
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
