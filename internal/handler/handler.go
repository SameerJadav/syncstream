package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/SameerJadav/syncstream/internal/logger"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type Room struct {
	Id      string
	Members [2]*websocket.Conn
}

var rooms = make(map[string]*Room)

func CreateRoom(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		logger.Error.Println("media type not supported")
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	var body struct{ VideoURL string }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Error.Printf("failed to decode JSON: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var start, end int
	for i := len(body.VideoURL) - 1; i >= 0; i-- {
		char := body.VideoURL[i]
		if char == '?' {
			end = i
		} else if char == '/' {
			start = i + 1
			break
		}
	}

	id := uuid.NewString()

	res := map[string]string{"pathname": fmt.Sprintf("/rooms/%s?videoid=%s", id, body.VideoURL[start:end])}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		logger.Error.Printf("failed to encode JSON: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func JoinRoom(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		logger.Error.Println("ID not found")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if _, err := uuid.Parse(id); err != nil {
		logger.Error.Printf("ID passed in URL path is not a valid UUID: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if room, ok := rooms[id]; ok {
		if room.Members[0] != nil && room.Members[1] != nil {
			logger.Error.Println("room is full")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		} else {
			renderRoom(w, r)
		}
	} else {
		rooms[id] = &Room{Id: id}
		renderRoom(w, r)
	}
}

func UpgradeConnection(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		logger.Error.Println("ID not found")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if _, err := uuid.Parse(id); err != nil {
		logger.Error.Printf("ID passed in URL path is not a valid UUID: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	room, ok := rooms[id]
	if !ok {
		logger.Error.Println("room does not exist")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		logger.Error.Printf("failed to upgrade connection: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var memberIdx uint8
	if room.Members[0] == nil {
		room.Members[0] = conn
	} else {
		room.Members[1] = conn
		memberIdx = 1
	}

	defer func() {
		conn.Close(websocket.StatusGoingAway, websocket.StatusGoingAway.String())
		room.Members[memberIdx] = nil
		if room.Members[0] == nil && room.Members[1] == nil {
			delete(rooms, id)
		}
	}()

	for {
		msgType, msg, err := conn.Read(context.Background())
		if err != nil {
			logger.Error.Printf("failed to read message: %v\n", err)
			return
		}

		for _, member := range room.Members {
			if member != nil && member != conn {
				if err := member.Write(context.Background(), msgType, msg); err != nil {
					logger.Error.Printf("failed to write message: %v\n", err)
					return
				}
			}
		}
	}
}

func renderRoom(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/room.html")
	if err != nil {
		logger.Error.Printf("failed to parse template file: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := struct{ VideoID string }{VideoID: r.URL.Query().Get("videoid")}
	if err := tmpl.Execute(w, data); err != nil {
		logger.Error.Printf("failed to parse template file: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
