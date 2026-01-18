package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mqtt-home/mqtt-lamarzocco/lamarzocco"
	"github.com/philipparndt/go-logger"
	loggerchi "github.com/philipparndt/go-logger-chi"
)

type SSEClient struct {
	ID      string
	Channel chan string
}

type WebServer struct {
	client        *lamarzocco.Client
	router        *chi.Mux
	sseClients    map[string]*SSEClient
	sseClientsMu  sync.RWMutex
	statusChan    chan lamarzocco.MachineStatus
}

type SetModeRequest struct {
	Mode string `json:"mode"`
}

type SetDoseRequest struct {
	DoseId string  `json:"doseId"`
	Dose   float64 `json:"dose"`
}

func NewWebServer(client *lamarzocco.Client) *WebServer {
	ws := &WebServer{
		client:     client,
		router:     chi.NewRouter(),
		sseClients: make(map[string]*SSEClient),
		statusChan: make(chan lamarzocco.MachineStatus, 10),
	}

	// Set callback to receive status updates
	originalCallback := client.SetStatusChangeCallback
	client.SetStatusChangeCallback(func(status lamarzocco.MachineStatus) {
		ws.onStatusChange(status)
	})
	_ = originalCallback // suppress unused warning

	ws.setupRoutes()
	go ws.broadcastLoop()

	return ws
}

func (ws *WebServer) onStatusChange(status lamarzocco.MachineStatus) {
	select {
	case ws.statusChan <- status:
	default:
		// Channel full, skip
	}
}

func (ws *WebServer) broadcastLoop() {
	for status := range ws.statusChan {
		ws.broadcastStatus(status)
	}
}

func (ws *WebServer) setupRoutes() {
	ws.router.Use(loggerchi.Middleware())
	ws.router.Use(middleware.Recoverer)

	ws.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	ws.router.Route("/api", func(r chi.Router) {
		r.Get("/health", ws.healthCheck)
		r.Get("/status", ws.getStatus)
		r.Post("/mode", ws.setMode)
		r.Post("/dose", ws.setDose)
		r.Post("/power", ws.setPower)
		r.Post("/backflush", ws.startBackFlush)
		r.Get("/events", ws.handleSSE)
	})

	// Serve static files (React app)
	fileServer := http.FileServer(http.Dir("./web/dist/"))
	ws.router.Handle("/*", fileServer)
}

func (ws *WebServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":     "ok",
		"goroutines": runtime.NumGoroutine(),
		"sse_clients": func() int {
			ws.sseClientsMu.RLock()
			defer ws.sseClientsMu.RUnlock()
			return len(ws.sseClients)
		}(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (ws *WebServer) getStatus(w http.ResponseWriter, r *http.Request) {
	status := ws.client.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (ws *WebServer) setMode(w http.ResponseWriter, r *http.Request) {
	var req SetModeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	mode := lamarzocco.ParseDoseMode(req.Mode)
	logger.Info("Setting mode via web API", "mode", mode)

	go func() {
		if err := ws.client.SetMode(mode); err != nil {
			logger.Error("Failed to set mode", "error", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (ws *WebServer) setDose(w http.ResponseWriter, r *http.Request) {
	var req SetDoseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate doseId
	if req.DoseId != "Dose1" && req.DoseId != "Dose2" {
		http.Error(w, "Invalid doseId, must be Dose1 or Dose2", http.StatusBadRequest)
		return
	}

	// Validate dose range
	if req.Dose < 5 || req.Dose > 100 {
		http.Error(w, "Dose must be between 5 and 100 grams", http.StatusBadRequest)
		return
	}

	logger.Info("Setting dose via web API", "doseId", req.DoseId, "dose", req.Dose)

	go func() {
		if err := ws.client.SetDose(req.DoseId, req.Dose); err != nil {
			logger.Error("Failed to set dose", "error", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

type SetPowerRequest struct {
	On bool `json:"on"`
}

func (ws *WebServer) setPower(w http.ResponseWriter, r *http.Request) {
	var req SetPowerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Info("Setting power via web API", "on", req.On)

	go func() {
		if err := ws.client.SetPower(req.On); err != nil {
			logger.Error("Failed to set power", "error", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (ws *WebServer) startBackFlush(w http.ResponseWriter, r *http.Request) {
	logger.Info("Starting back flush via web API")

	go func() {
		if err := ws.client.StartBackFlush(); err != nil {
			logger.Error("Failed to start back flush", "error", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (ws *WebServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientID := fmt.Sprintf("%d", time.Now().UnixNano())
	logger.Info("SSE client connected", "id", clientID)

	channel := make(chan string, 10)

	ws.sseClientsMu.Lock()
	ws.sseClients[clientID] = &SSEClient{
		ID:      clientID,
		Channel: channel,
	}
	ws.sseClientsMu.Unlock()

	// Send initial state
	status := ws.client.GetStatus()
	message, _ := json.Marshal(status)
	fmt.Fprintf(w, "data: %s\n\n", string(message))

	flusher, ok := w.(http.Flusher)
	if ok {
		flusher.Flush()
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	defer func() {
		logger.Info("SSE client disconnected", "id", clientID)
		ws.sseClientsMu.Lock()
		delete(ws.sseClients, clientID)
		close(channel)
		ws.sseClientsMu.Unlock()
	}()

	for {
		select {
		case msg := <-channel:
			_, writeErr := fmt.Fprintf(w, "data: %s\n\n", msg)
			if writeErr != nil {
				return
			}
			if ok {
				flusher.Flush()
			}
		case <-r.Context().Done():
			return
		case <-ticker.C:
			status := ws.client.GetStatus()
			message, _ := json.Marshal(status)
			_, writeErr := fmt.Fprintf(w, "data: %s\n\n", string(message))
			if writeErr != nil {
				return
			}
			if ok {
				flusher.Flush()
			}
		}
	}
}

func (ws *WebServer) broadcastStatus(status lamarzocco.MachineStatus) {
	message, err := json.Marshal(status)
	if err != nil {
		logger.Error("Failed to marshal status", "error", err)
		return
	}
	messageStr := string(message)

	ws.sseClientsMu.RLock()
	for _, client := range ws.sseClients {
		select {
		case client.Channel <- messageStr:
		default:
			// Channel full, skip
		}
	}
	ws.sseClientsMu.RUnlock()
}

func (ws *WebServer) Start(port int) error {
	addr := ":" + strconv.Itoa(port)
	logger.Info("Starting web server", "address", addr)
	return http.ListenAndServe(addr, ws.router)
}
