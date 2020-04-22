package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func submit(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	_, bytes, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}
	runContainer(string(bytes), conn)
}

func server() {
	http.HandleFunc("/submit", submit)
	log.Fatal(http.ListenAndServe(":7867", nil))
}
