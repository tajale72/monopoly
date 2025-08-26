package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// broadcast sends the same message to every client in the room.
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

/* ===== Models ===== */

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

type rollReq struct {
	PlayerID string `json:"playerId"`
	Room     string `json:"room"`
	Name     string `json:"name"` // optional (used for logging)
}

/* ===== Globals ===== */

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	mu    sync.Mutex
	rooms = make(map[string]map[*Client]struct{}) // room -> clients
	turn  = make(map[string]*Client)              // room -> current player

	// positions[room][playerID] = tileIndex (0..39). We keep server-authoritative positions.
	positions = make(map[string]map[string]int)

	maxPlayers = 10
)

/* ===== CORS ===== */

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}

/* ===== Main ===== */

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/ws", withCORS(wsHandler))
	http.HandleFunc("/roll", withCORS(rollHTTP))
	http.HandleFunc("/debug/players", withCORS(debugPlayersHTTP))

	// Serve HTML
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/", "/index.html":
			http.ServeFile(w, r, "index.html")
			return
		case "/game.html":
			http.ServeFile(w, r, "game.html")
			return
		default:
			http.NotFound(w, r)
			return
		}
	})

	addr := ":8080"
	log.Printf("listening on %s (ws: /ws, roll: /roll)", addr)
	go broadcastServerLog("Server started; waiting for players...")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

/* ===== REST: /roll ===== */

func rollHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("HTTP /roll", r.RemoteAddr)
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req rollReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("decode error: %v", err)
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if req.Room == "" || req.PlayerID == "" {
		http.Error(w, "missing room or playerId", http.StatusBadRequest)
		return
	}

	c := getClientByID(req.Room, req.PlayerID)
	if c == nil {
		http.Error(w, "player not connected in room", http.StatusNotFound)
		return
	}

	// Turn check
	holder := getTurnHolder(req.Room)
	if holder == nil || holder != c {
		http.Error(w, "not your turn", http.StatusForbidden)
		c.writeJSON(map[string]any{"type": "event", "text": "Not your turn."})
		return
	}

	// Authoritative dice & move
	d1, d2 := 1+rand.Intn(6), 1+rand.Intn(6)
	total := d1 + d2

	from := getPos(req.Room, c.ID) // defaults to 0 if absent
	to := (from + total) % 40
	setPos(req.Room, c.ID, to)

	// Broadcast event + move (with dice for the roller)
	broadcast(req.Room, map[string]any{
		"type": "event",
		"text": fmt.Sprintf("%s rolled %d (%d + %d)", c.Name, total, d1, d2),
	})
	broadcast(req.Room, map[string]any{
		"type":     "move",
		"playerId": c.ID,
		"from":     from,
		"to":       to,
		"dice":     []int{d1, d2},
	})

	// Pass turn
	passTurn(req.Room, c)

	// Response for the caller
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":    true,
		"dice":  []int{d1, d2},
		"total": total,
	})
}

/* ===== WS: /ws ===== */

func wsHandler(w http.ResponseWriter, r *http.Request) {
	cn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	client := &Client{Conn: cn}

	defer func() {
		onClose(client)
		_ = cn.Close()
	}()

	for {
		_, data, err := cn.ReadMessage()
		if err != nil {
			return
		}

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
				client.writeJSON(map[string]any{"type": "event", "text": "Room is full (10 players max)."})
				return
			}
			broadcastServerLogTo(client.Room, fmt.Sprintf("%s connected (%s)", client.Name, short(client.ID)))

			// Ensure initial position at GO
			ensurePos(client.Room, client.ID)

			// Broadcast roster + joined delta
			broadcast(client.Room, map[string]any{"type": "players", "list": roster(client.Room)})
			broadcast(client.Room, map[string]any{"type": "playerJoined", "player": Player{ID: client.ID, Name: client.Name}})

			// Send a state snapshot so clients can render tokens (GO for new players)
			broadcast(client.Room, map[string]any{"type": "state", "positions": snapshotPositions(client.Room)})

			// Ensure someone has the turn
			ensureTurnHolder(client.Room)

		case "who":
			room := in.Room
			if room == "" {
				room = client.Room
			}
			client.writeJSON(map[string]any{"type": "players", "list": roster(room)})

		case "subscribeLogs":
			client.writeJSON(map[string]any{"type": "serverLog", "text": "Subscribed to server logs for room " + client.Room})

		case "roll":
			room := client.Room
			if room == "" {
				break
			}
			holder := getTurnHolder(room)
			if holder == nil || holder != client {
				client.writeJSON(map[string]any{"type": "event", "text": "Not your turn."})
				break
			}
			// Server-authoritative dice + move
			d1, d2 := 1+rand.Intn(6), 1+rand.Intn(6)
			total := d1 + d2

			from := getPos(room, client.ID)
			to := (from + total) % 40
			setPos(room, client.ID, to)

			broadcast(room, map[string]any{
				"type": "event",
				"text": fmt.Sprintf("%s rolled %d (%d + %d)", client.Name, total, d1, d2),
			})
			broadcast(room, map[string]any{
				"type":     "move",
				"playerId": client.ID,
				"from":     from,
				"to":       to,
				"dice":     []int{d1, d2},
			})

			passTurn(room, client)

		case "ping":
			// ignore

		case "leave":
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

		// Update roster + left delta
		broadcast(c.Room, map[string]any{"type": "players", "list": roster(c.Room)})
		broadcast(c.Room, map[string]any{"type": "playerLeft", "player": Player{ID: c.ID, Name: c.Name}})

		// If turn holder left, advance
		holder := getTurnHolder(c.Room)
		if holder == nil || holder == c {
			passTurn(c.Room, c)
		}
	}
}

/* ===== Rooms / Roster ===== */

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
		delete(positions, room)
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
	// Sort for stable order (by Name then ID)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name == out[j].Name {
			return out[i].ID < out[j].ID
		}
		return out[i].Name < out[j].Name
	})
	return out
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

/* ===== Positions (server-authoritative) ===== */

func ensurePos(room, playerID string) {
	mu.Lock()
	defer mu.Unlock()
	if positions[room] == nil {
		positions[room] = make(map[string]int)
	}
	if _, ok := positions[room][playerID]; !ok {
		positions[room][playerID] = 0 // GO for brand new players
	}
}

func getPos(room, playerID string) int {
	mu.Lock()
	defer mu.Unlock()
	if positions[room] == nil {
		return 0
	}
	return positions[room][playerID]
}

func setPos(room, playerID string, idx int) {
	mu.Lock()
	defer mu.Unlock()
	if positions[room] == nil {
		positions[room] = make(map[string]int)
	}
	positions[room][playerID] = ((idx % 40) + 40) % 40
}

func snapshotPositions(room string) map[string]int {
	mu.Lock()
	defer mu.Unlock()
	cp := make(map[string]int, len(positions[room]))
	for k, v := range positions[room] {
		cp[k] = v
	}
	return cp
}

/* ===== Turns ===== */

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

	// Build deterministic list (sort by Name, then ID) so rotation is stable
	list := make([]*Client, 0, len(set))
	for c := range set {
		list = append(list, c)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Name == list[j].Name {
			return list[i].ID < list[j].ID
		}
		return list[i].Name < list[j].Name
	})

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
	broadcast(room, map[string]any{"type": "event", "text": fmt.Sprintf("It's %s's turn.", holder.Name)})
	holder.writeJSON(map[string]any{"type": "yourTurn", "canRoll": true})
}

/* ===== Logging ===== */

func broadcastServerLog(text string) {
	log.Println(text)
	mu.Lock()
	var list []string
	for room := range rooms {
		list = append(list, room)
	}
	mu.Unlock()
	for _, room := range list {
		broadcast(room, map[string]any{"type": "serverLog", "text": text})
	}
}

func broadcastServerLogTo(room, text string) {
	log.Printf("[%s] %s", room, text)
	broadcast(room, map[string]any{"type": "serverLog", "text": text})
}

/* ===== Client write helpers ===== */

func (c *Client) writeJSON(v any) {
	b, _ := json.Marshal(v)
	c.writeRaw(b)
}

func (c *Client) writeRaw(b []byte) {
	c.SendMu.Lock()
	defer c.SendMu.Unlock()
	_ = c.Conn.WriteMessage(websocket.TextMessage, b)
}

/* ===== Utils ===== */

func short(id string) string {
	if len(id) <= 6 {
		return id
	}
	return id[:6]
}

/* ===== Debug endpoint ===== */

type PlayerInfo struct {
	PlayerID string `json:"playerId"`
	Name     string `json:"name"`
	RoomID   string `json:"roomId"`
	Pos      int    `json:"pos"`
}

func debugPlayersHTTP(w http.ResponseWriter, r *http.Request) {
	room := r.URL.Query().Get("room")
	w.Header().Set("Content-Type", "application/json")

	mu.Lock()
	defer mu.Unlock()

	if room != "" {
		var list []PlayerInfo
		set := rooms[room]
		for c := range set {
			pos := 0
			if positions[room] != nil {
				pos = positions[room][c.ID]
			}
			list = append(list, PlayerInfo{
				PlayerID: c.ID,
				Name:     c.Name,
				RoomID:   room,
				Pos:      pos,
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"room": room, "players": list})
		return
	}

	out := map[string][]PlayerInfo{}
	for rm, set := range rooms {
		for c := range set {
			pos := 0
			if positions[rm] != nil {
				pos = positions[rm][c.ID]
			}
			out[rm] = append(out[rm], PlayerInfo{
				PlayerID: c.ID,
				Name:     c.Name,
				RoomID:   rm,
				Pos:      pos,
			})
		}
	}
	_ = json.NewEncoder(w).Encode(out)
}
