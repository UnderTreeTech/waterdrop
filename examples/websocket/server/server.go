/*
 *
 * Copyright 2020 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http/server"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
	ws "github.com/UnderTreeTech/waterdrop/pkg/server/http/websocket"
	"github.com/gorilla/websocket"
)

var pong = []byte("pong")

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	defer log.New(nil).Sync()

	srv := server.New(nil)
	srv.Upgrade(ws.NewWebSocket("/ws", func(ws *ws.WebSocket) {
		ws.SetPingHandler(func(message string) error {
			ws.SetReadDeadline(time.Now().Add(time.Second * 10))
			return ws.WriteControl(websocket.PongMessage, pong, time.Now().Add(time.Second))
		})
		for {
			ws.SetReadDeadline(time.Now().Add(time.Second * 10))
			msgType, message, err := ws.ReadMessage()
			if err != nil {
				fmt.Println("read msg fail", err.Error())
				break
			}
			fmt.Println("recv msg", string(message), msgType)

			err = ws.WriteMessage(msgType, message)
			if err != nil {
				fmt.Println("write msg fail", err.Error())
				break
			}
		}
	}))

	srv.Start()

	<-c
}
