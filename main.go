package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Client struct {
	ID, Name, Room string
	Conn           *websocket.Conn
	SendMu         sync.Mutex // serialize writes
}

type inbound struct {
	Type     string `json:"type"`
	PlayerID string `json:"playerId"`
	Name     string `json:"name"`
	Room     string `json:"room"`
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	mu    sync.Mutex
	rooms = make(map[string]map[*Client]struct{}) // room -> clients

	// Optional: simple turn management per room
	turn = make(map[string]*Client)

	maxPlayers = 10
)

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/roll", rollHTTP)

	// serve game.html and static
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/game.html" {
			http.ServeFile(w, r, "game.html")
			return
		}
		http.NotFound(w, r)
	})

	addr := ":8080"
	log.Printf("listening on %s (ws endpoint: /ws)", addr)
	go broadcastServerLog("Server started; waiting for players...")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

type rollReq struct {
	PlayerID string `json:"playerId"`
	Room     string `json:"room"`
}

func rollHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Rolling Dice", r.RemoteAddr)
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req rollReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if req.Room == "" || req.PlayerID == "" {
		http.Error(w, "missing room or playerId", http.StatusBadRequest)
		return
	}

	// Resolve the client
	c := getClientByID(req.Room, req.PlayerID)
	if c == nil {
		http.Error(w, "player not connected in room", http.StatusNotFound)
		return
	}

	// Verify turn holder
	holder := getTurnHolder(req.Room)
	if holder == nil || holder != c {
		http.Error(w, "not your turn", http.StatusForbidden)
		// Optional: also tell the player via WS
		if c != nil {
			c.writeJSON(map[string]any{
				"type": "event",
				"text": "Not your turn.",
			})
		}
		return
	}

	// Roll dice
	d1 := 1 + rand.Intn(6)
	d2 := 1 + rand.Intn(6)
	total := d1 + d2

	// Broadcast event text
	broadcast(req.Room, map[string]any{
		"type": "event",
		"text": fmt.Sprintf("%s rolled %d (%d + %d)", c.Name, total, d1, d2),
	})

	// (Optional) Animate a token move if you track positions.
	// If you keep per-player positions server-side, compute from->to and emit:
	// broadcast(req.Room, map[string]any{
	// 	"type":     "move",
	// 	"playerId": c.ID,
	// 	"from":     fromIndex,
	// 	"to":       (fromIndex + total) % 40,
	// })

	// Pass the turn and notify next holder
	passTurn(req.Room, c)

	// Respond to the HTTP caller
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":    true,
		"dice":  []int{d1, d2},
		"total": total,
	})
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	cn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	client := &Client{Conn: cn}

	defer func() {
		onClose(client)
		cn.Close()
	}()

	for {
		_, data, err := cn.ReadMessage()
		if err != nil {
			return
		}

		// echo raw inbound to any log-subscribers in the room (after we know room)
		var msg map[string]any
		_ = json.Unmarshal(data, &msg)

		// handle typed messages
		var in inbound
		_ = json.Unmarshal(data, &in)

		switch in.Type {
		case "resume":
			client.ID = in.PlayerID
			client.Name = in.Name
			if in.Room != "" {
				client.Room = in.Room
			} else if client.Room == "" {
				client.Room = "default"
			}

			if !addToRoom(client.Room, client) {
				// room full
				client.writeJSON(map[string]any{
					"type": "event",
					"text": "Room is full (10 players max).",
				})
				return
			}

			broadcastServerLogTo(client.Room, fmt.Sprintf("%s connected (%s)", client.Name, short(client.ID)))

			// snapshot roster to everyone
			broadcast(client.Room, map[string]any{
				"type": "players",
				"list": roster(client.Room),
			})
			// delta joined
			broadcast(client.Room, map[string]any{
				"type":   "playerJoined",
				"player": Player{ID: client.ID, Name: client.Name},
			})

			// If no one has the turn yet, give it to first player
			ensureTurnHolder(client.Room)

		case "who":
			room := in.Room
			if room == "" {
				room = client.Room
			}
			client.writeJSON(map[string]any{
				"type": "players",
				"list": roster(room),
			})

		case "subscribeLogs":
			// No explicit subscription list; we just broadcast important logs to the room.
			// This message is here so the client knows it's set up.
			client.writeJSON(map[string]any{
				"type": "serverLog",
				"text": "Subscribed to server logs for room " + client.Room,
			})

		case "roll":
			// simple demo roll, emits an event, and passes turn
			room := client.Room
			if room == "" {
				break
			}
			holder := getTurnHolder(room)
			if holder == nil || holder != client {
				client.writeJSON(map[string]any{
					"type": "event",
					"text": "Not your turn.",
				})
				break
			}

			d1 := 1 + rand.Intn(6)
			d2 := 1 + rand.Intn(6)
			total := d1 + d2
			broadcast(room, map[string]any{
				"type": "event",
				"text": fmt.Sprintf("%s rolled %d (%d + %d)", client.Name, total, d1, d2),
			})

			// pass turn to next connected player
			passTurn(room, client)

		case "ping":
			// ignore

		case "leave":
			// client initiated close
			return
		}
	}
}

func onClose(c *Client) {
	if c.Room == "" {
		return
	}
	removed := removeFromRoom(c.Room, c)
	if removed {
		broadcastServerLogTo(c.Room, fmt.Sprintf("%s disconnected (%s)", c.Name, short(c.ID)))

		// send fresh roster + delta left
		broadcast(c.Room, map[string]any{
			"type": "players",
			"list": roster(c.Room),
		})
		broadcast(c.Room, map[string]any{
			"type":   "playerLeft",
			"player": Player{ID: c.ID, Name: c.Name},
		})

		// if the turn holder left, move turn forward
		holder := getTurnHolder(c.Room)
		if holder == nil || holder == c {
			passTurn(c.Room, c)
		}
	}
}

// ---- rooms/roster helpers ----

func addToRoom(room string, c *Client) bool {
	mu.Lock()
	defer mu.Unlock()

	set := rooms[room]
	if set == nil {
		set = make(map[*Client]struct{})
		rooms[room] = set
	}
	if len(set) >= maxPlayers {
		return false
	}
	set[c] = struct{}{}
	return true
}

func removeFromRoom(room string, c *Client) bool {
	mu.Lock()
	defer mu.Unlock()
	set := rooms[room]
	if set == nil {
		return false
	}
	if _, ok := set[c]; !ok {
		return false
	}
	delete(set, c)
	if len(set) == 0 {
		delete(rooms, room)
		delete(turn, room)
	}
	return true
}

func roster(room string) []Player {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Player, 0, 8)
	for cl := range rooms[room] {
		out = append(out, Player{ID: cl.ID, Name: cl.Name})
	}
	return out
}

func broadcast(room string, msg any) {
	b, _ := json.Marshal(msg)

	mu.Lock()
	set := rooms[room]
	clients := make([]*Client, 0, len(set))
	for cl := range set {
		clients = append(clients, cl)
	}
	mu.Unlock()

	for _, cl := range clients {
		cl.writeRaw(b)
	}
}

// ---- simple turn handling ----

func ensureTurnHolder(room string) {
	mu.Lock()
	defer mu.Unlock()
	if turn[room] != nil {
		return
	}
	for cl := range rooms[room] {
		turn[room] = cl
		break
	}
	if turn[room] != nil {
		go notifyTurn(room)
	}
}

func getTurnHolder(room string) *Client {
	mu.Lock()
	defer mu.Unlock()
	return turn[room]
}

func passTurn(room string, current *Client) {
	mu.Lock()
	defer mu.Unlock()

	set := rooms[room]
	if len(set) == 0 {
		turn[room] = nil
		return
	}

	// build stable list to rotate
	list := make([]*Client, 0, len(set))
	for c := range set {
		list = append(list, c)
	}
	// find current index
	nextIdx := 0
	for i, c := range list {
		if c == current {
			nextIdx = (i + 1) % len(list)
			break
		}
	}
	turn[room] = list[nextIdx]
	go notifyTurn(room)
}

func notifyTurn(room string) {
	holder := getTurnHolder(room)
	if holder == nil {
		return
	}
	// notify everyone and specifically enable for holder
	broadcast(room, map[string]any{
		"type": "event",
		"text": fmt.Sprintf("It's %s's turn.", holder.Name),
	})
	// tell holder they can roll
	holder.writeJSON(map[string]any{
		"type":    "yourTurn",
		"canRoll": true,
	})
}

// ---- logging helpers (mirror to clients via serverLog) ----

func broadcastServerLog(text string) {
	log.Println(text)
	// send to all rooms
	mu.Lock()
	var roomsList []string
	for room := range rooms {
		roomsList = append(roomsList, room)
	}
	mu.Unlock()
	for _, room := range roomsList {
		broadcast(room, map[string]any{"type": "serverLog", "text": text})
	}
}

func broadcastServerLogTo(room, text string) {
	log.Printf("[%s] %s", room, text)
	broadcast(room, map[string]any{"type": "serverLog", "text": text})
}

// ---- client write helpers ----

func (c *Client) writeJSON(v any) {
	b, _ := json.Marshal(v)
	c.writeRaw(b)
}

func (c *Client) writeRaw(b []byte) {
	c.SendMu.Lock()
	defer c.SendMu.Unlock()
	_ = c.Conn.WriteMessage(websocket.TextMessage, b)
}

func short(id string) string {
	if len(id) <= 6 {
		return id
	}
	return id[:6]
}

func getClientByID(room, playerID string) *Client {
	mu.Lock()
	defer mu.Unlock()
	for c := range rooms[room] {
		if c.ID == playerID {
			return c
		}
	}
	return nil
}
