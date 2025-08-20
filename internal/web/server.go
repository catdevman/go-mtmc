package web

import (
	"encoding/json"
	"github.com/catdevman/go-mtmc/internal/emulator"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Server holds the dependencies for the web server.
type Server struct {
	computer  *emulator.MonTanaMiniComputer
	templates map[string]*template.Template
}

// NewServer creates a new web server.
func NewServer(computer *emulator.MonTanaMiniComputer) *Server {
	s := &Server{
		computer:  computer,
		templates: make(map[string]*template.Template),
	}
	s.parseTemplates()
	return s
}

func (s *Server) parseTemplates() {
	s.templates["index"] = template.Must(template.ParseFiles("web/templates/index.html", "web/templates/layout.html"))
}

// Start begins listening for HTTP requests.
func (s *Server) Start() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/ws", s.handleWebSocket)
	http.HandleFunc("/control", s.handleControl)

	log.Println("Starting web server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	err := s.templates["index"].ExecuteTemplate(w, "layout", s.computer.GetState())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// Register the WebSocket connection as an observer
	observer := &WebSocketObserver{conn: conn}
	s.computer.AddObserver(observer)

	// Keep the connection alive
	for {
		// The read loop is just to keep the connection open.
		// All updates are pushed from the computer via the observer interface.
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}
}

func (s *Server) handleControl(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	switch action {
	case "run":
		s.computer.Running = true
	case "pause":
		s.computer.Running = false
	case "step":
		s.computer.Step()
	case "reset":
		// A more complete reset would be needed here
		s.computer.PC = 0
		s.computer.Running = false
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// WebSocketObserver sends computer state updates to a WebSocket client.
type WebSocketObserver struct {
	conn *websocket.Conn
}

// Update sends the computer's state to the WebSocket client.
func (o *WebSocketObserver) Update(computer *emulator.MonTanaMiniComputer) {
	state := computer.GetState()
	data, err := json.Marshal(state)
	if err != nil {
		log.Println("Error marshalling state:", err)
		return
	}
	if err := o.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		// Client has likely disconnected
	}
}
