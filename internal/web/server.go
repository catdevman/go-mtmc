package web

import (
	"embed"
	"encoding/json"
	"github.com/catdevman/go-mtmc/internal/disk"
	"github.com/catdevman/go-mtmc/internal/emulator"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

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
	s.templates["index"] = template.Must(template.ParseFS(templatesFS, "templates/index.html", "templates/layout.html"))
}

// Start begins listening for HTTP requests.
func (s *Server) Start() {
	staticContent, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticContent))))

	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/ws", s.handleWebSocket)
	http.HandleFunc("/control", s.handleControl)
	http.HandleFunc("/load", s.handleLoad)

	log.Println("Starting web server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	files, err := fs.ReadDir(disk.FS, "disk/bin")
	if err != nil {
		http.Error(w, "could not read programs directory", http.StatusInternalServerError)
		return
	}

	var programs []string
	for _, file := range files {
		programs = append(programs, file.Name())
	}

	data := s.computer.GetState()
	data["programs"] = programs

	err = s.templates["index"].ExecuteTemplate(w, "layout", data)
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
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}
}

func (s *Server) handleControl(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	switch action {
	case "run":
		log.Println("sent action run")
		s.computer.Running = true
	case "pause":
		log.Println("sent action pause")
		s.computer.Running = false
	case "step":
		log.Println("sent action step")
		s.computer.Step()
	case "reset":
		log.Println("sent action reset")
		s.computer.Registers[emulator.PC] = 0
		s.computer.Running = false
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handleLoad(w http.ResponseWriter, r *http.Request) {
	programName := r.URL.Query().Get("program")
	if programName == "" {
		http.Error(w, "program name is required", http.StatusBadRequest)
		return
	}

	program, err := fs.ReadFile(disk.FS, "disk/bin/"+programName)
	if err != nil {
		http.Error(w, "could not read program", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	s.computer.LoadProgram(program, 0)
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
