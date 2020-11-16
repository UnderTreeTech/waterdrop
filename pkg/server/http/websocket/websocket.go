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

package websocket

import (
	"net/http"

	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/gorilla/websocket"
)

type WebSocketHandler func(*WebSocket)

type WebSocket struct {
	Path    string
	Handler WebSocketHandler

	*websocket.Upgrader
	*websocket.Conn
}

func NewWebSocket(path string, handler WebSocketHandler) *WebSocket {
	return &WebSocket{
		Path:     path,
		Upgrader: &websocket.Upgrader{},
		Handler:  handler,
	}
}

func (ws *WebSocket) Upgrade(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(r.Context(), "upgrade fail", log.String("error", err.Error()))
		return
	}
	defer conn.Close()

	ws.Conn = conn
	ws.Handler(ws)
}
