package daemon

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"ai-manager/internal/backup"
	"ai-manager/internal/cleanup"
	"ai-manager/internal/config"
	"ai-manager/internal/credentials"
	"ai-manager/internal/discovery"
	"ai-manager/internal/link"
	"ai-manager/internal/models"
	"ai-manager/internal/scheduler"
	"ai-manager/internal/switcher"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	// ServerAddr is the default address for the daemon
	ServerAddr = "127.0.0.1:19999"
	// WebSocketPingInterval is the interval for ping/pong
	WebSocketPingInterval = 30 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Server represents the daemon server
type Server struct {
	router     *mux.Router
	cfg        *config.Config
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	stopCh     chan struct{}
	httpServer *http.Server
}

// NewServer creates a new daemon server
func NewServer(cfg *config.Config) *Server {
	return &Server{
		router:   mux.NewRouter(),
		cfg:      cfg,
		clients:  make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte, 256),
		stopCh:   make(chan struct{}),
	}
}

// Start starts the daemon server
func (s *Server) Start(addr string) error {
	if addr == "" {
		addr = ServerAddr
	}

	// Setup routes
	s.setupRoutes()

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      handlers.CORS()(s.router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go s.handleWebSocket()
	go s.handleShutdown(sigCh)

	log.Printf("Starting daemon server on %s", addr)
	return s.httpServer.ListenAndServe()
}

// Stop stops the daemon server
func (s *Server) Stop() error {
	close(s.stopCh)

	// Close all WebSocket clients
	for client := range s.clients {
		client.Close()
	}

	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/api/health", s.healthHandler).Methods("GET")

	// API routes
	s.router.HandleFunc("/api/tools", s.toolsHandler).Methods("GET")
	s.router.HandleFunc("/api/tools/{name}/stats", s.toolStatsHandler).Methods("GET")
	s.router.HandleFunc("/api/tools/{name}/cleanup", s.toolCleanupHandler).Methods("POST")

	s.router.HandleFunc("/api/models", s.modelsHandler).Methods("GET")
	s.router.HandleFunc("/api/switch", s.switchModelHandler).Methods("POST")

	s.router.HandleFunc("/api/backups", s.backupsHandler).Methods("GET")
	s.router.HandleFunc("/api/backups", s.createBackupHandler).Methods("POST")
	s.router.HandleFunc("/api/backups/{id}/restore", s.restoreBackupHandler).Methods("POST")
	s.router.HandleFunc("/api/backups/{id}", s.deleteBackupHandler).Methods("DELETE")

	s.router.HandleFunc("/api/links", s.linksHandler).Methods("GET")
	s.router.HandleFunc("/api/links", s.createLinkHandler).Methods("POST")
	s.router.HandleFunc("/api/links/{name}", s.deleteLinkHandler).Methods("DELETE")

	s.router.HandleFunc("/api/stats", s.statsHandler).Methods("GET")

	// Scheduler endpoints
	s.router.HandleFunc("/api/scheduler", s.schedulerHandler).Methods("GET")
	s.router.HandleFunc("/api/scheduler", s.schedulerCreateHandler).Methods("POST")
	s.router.HandleFunc("/api/scheduler/{id}", s.schedulerUpdateHandler).Methods("PUT")
	s.router.HandleFunc("/api/scheduler/{id}", s.schedulerDeleteHandler).Methods("DELETE")
	s.router.HandleFunc("/api/scheduler/{id}/run", s.schedulerRunHandler).Methods("POST")

	// Credentials endpoints
	s.router.HandleFunc("/api/credentials", s.credentialsHandler).Methods("GET")
	s.router.HandleFunc("/api/credentials", s.credentialsSetHandler).Methods("POST")
	s.router.HandleFunc("/api/credentials/{model}/{key}", s.credentialsDeleteHandler).Methods("DELETE")
	s.router.HandleFunc("/api/credentials/{model}", s.modelCredentialsHandler).Methods("GET")

	// WebSocket endpoint
	s.router.HandleFunc("/ws", s.wsHandler)

	// Serve static files (for bundled UI)
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./ui/dist")))
}

// Health check handler
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "0.2.0",
	})
}

// Tools handlers
func (s *Server) toolsHandler(w http.ResponseWriter, r *http.Request) {
	scanner := discovery.NewScanner(s.cfg)
	result, err := scanner.Scan()
	if err != nil {
		sendError(w, "Failed to scan tools: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, result)
}

func (s *Server) toolStatsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	scanner := discovery.NewScanner(s.cfg)
	result, err := scanner.Scan()
	if err != nil {
		sendError(w, "Failed to scan tools: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, tool := range result.Tools {
		if tool.Name == name {
			sendJSON(w, tool)
			return
		}
	}

	sendError(w, "Tool not found: "+name, http.StatusNotFound)
}

func (s *Server) toolCleanupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	cleaner := cleanup.NewCleaner(s.cfg)
	result := models.CleanupResult{}

	// Find the tool in config
	for key, tool := range s.cfg.Tools {
		if tool.Name == name || key == name {
			result = cleaner.CleanupTool(key, tool)
			break
		}
	}

	if result.Tool == "" {
		sendError(w, "Tool not found: "+name, http.StatusNotFound)
		return
	}

	s.broadcastMessage(map[string]interface{}{
		"type":   "cleanup_complete",
		"tool":   name,
		"result": result,
	})

	sendJSON(w, result)
}

// Models handlers
func (s *Server) modelsHandler(w http.ResponseWriter, r *http.Request) {
	models := make([]map[string]interface{}, 0)
	for name, m := range s.cfg.Models {
		models = append(models, map[string]interface{}{
			"name":      name,
			"full_name": m.Name,
			"provider":  m.Provider,
			"endpoint":  m.APIEndpoint,
			"model_id":  m.ModelID,
			"default":   s.cfg.Defaults.Model == name,
		})
	}

	sendJSON(w, map[string]interface{}{
		"models":  models,
		"current": s.cfg.Defaults.Model,
	})
}

func (s *Server) switchModelHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	switcher := switcher.NewSwitcher(s.cfg)
	result, err := switcher.PerformSwitch(req.Model, false)
	if err != nil {
		sendError(w, "Failed to switch model: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.cfg.Defaults.Model = req.Model
	if err := config.Save(s.cfg, config.GetDefaultConfigPath()); err != nil {
		log.Printf("Warning: failed to save config: %v", err)
	}

	s.broadcastMessage(map[string]interface{}{
		"type":   "model_switched",
		"model":  req.Model,
		"result": result,
	})

	sendJSON(w, result)
}

// Backup handlers
func (s *Server) backupsHandler(w http.ResponseWriter, r *http.Request) {
	bm := backup.NewBackupManager(s.cfg)
	backups, err := bm.ListBackups()
	if err != nil {
		sendError(w, "Failed to list backups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]interface{}{
		"backups": backups,
	})
}

func (s *Server) createBackupHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	bm := backup.NewBackupManager(s.cfg)
	result, err := bm.Backup(true)
	if err != nil {
		sendError(w, "Failed to create backup: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.broadcastMessage(map[string]interface{}{
		"type":   "backup_created",
		"backup": result,
	})

	sendJSON(w, result)
}

func (s *Server) restoreBackupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	bm := backup.NewBackupManager(s.cfg)
	result, err := bm.Restore(id, "")
	if err != nil {
		sendError(w, "Failed to restore backup: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.broadcastMessage(map[string]interface{}{
		"type":   "backup_restored",
		"backup": id,
	})

	sendJSON(w, result)
}

func (s *Server) deleteBackupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	bm := backup.NewBackupManager(s.cfg)
	if err := bm.DeleteBackup(id); err != nil {
		sendError(w, "Failed to delete backup: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.broadcastMessage(map[string]interface{}{
		"type": "backup_deleted",
		"id":   id,
	})

	sendJSON(w, map[string]string{"status": "deleted"})
}

// Link handlers
func (s *Server) linksHandler(w http.ResponseWriter, r *http.Request) {
	lm := link.NewLinkManager(s.cfg)
	links, err := lm.ListLinks()
	if err != nil {
		sendError(w, "Failed to list links: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]interface{}{
		"links": links,
	})
}

func (s *Server) createLinkHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Tool string `json:"tool"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	lm := link.NewLinkManager(s.cfg)
	result, err := lm.CreateLink(req.Tool)
	if err != nil {
		sendError(w, "Failed to create link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.broadcastMessage(map[string]interface{}{
		"type":   "link_created",
		"tool":   req.Tool,
		"result": result,
	})

	sendJSON(w, result)
}

func (s *Server) deleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	lm := link.NewLinkManager(s.cfg)
	result, err := lm.RemoveLink(name)
	if err != nil {
		sendError(w, "Failed to remove link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.broadcastMessage(map[string]interface{}{
		"type": "link_removed",
		"name": name,
	})

	sendJSON(w, result)
}

// Stats handler
func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) {
	bm := backup.NewBackupManager(s.cfg)

	// Get disk usage
	var totalUsage int64
	for _, tool := range s.cfg.Tools {
		home, _ := os.UserHomeDir()
		basePath := tool.Path
		if basePath == "" {
			continue
		}
		if basePath[0] == '~' {
			basePath = filepath.Join(home, basePath[2:])
		} else {
			basePath = os.ExpandEnv(basePath)
		}
		if info, err := os.Stat(basePath); err == nil && info.IsDir() {
			totalUsage += getDirSize(basePath)
		}
	}

	backups, _ := bm.ListBackups()

	stats := map[string]interface{}{
		"disk_usage": map[string]interface{}{
			"total_bytes": totalUsage,
			"formatted":   models.FormatBytes(totalUsage),
		},
		"backups_count": len(backups),
		"tools_count":   len(s.cfg.Tools),
	}

	sendJSON(w, stats)
}

func getDirSize(path string) int64 {
	var size int64
	//nolint:errcheck
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// WebSocket handlers
func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	s.clients[conn] = true
	log.Printf("WebSocket client connected, total: %d", len(s.clients))

	// Start ping/pong
	go s.pingClient(conn)

	// Handle incoming messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Process incoming message (e.g., keep-alive)
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			s.handleWSMessage(conn, msg)
		}
	}

	delete(s.clients, conn)
	conn.Close()
	log.Printf("WebSocket client disconnected, total: %d", len(s.clients))
}

func (s *Server) handleWSMessage(conn *websocket.Conn, msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "ping":
		s.sendWSMessage(conn, map[string]interface{}{
			"type": "pong",
			"time": time.Now().Unix(),
		})
	case "scan":
		scanner := discovery.NewScanner(s.cfg)
		result, _ := scanner.Scan()
		s.sendWSMessage(conn, map[string]interface{}{
			"type":   "scan_result",
			"result": result,
		})
	}
}

func (s *Server) pingClient(conn *websocket.Conn) {
	ticker := time.NewTicker(WebSocketPingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				conn.Close()
				return
			}
		case <-s.stopCh:
			return
		}
	}
}

func (s *Server) handleWebSocket() {
	for {
		select {
		case msg := <-s.broadcast:
			for client := range s.clients {
				if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
					client.Close()
					delete(s.clients, client)
				}
			}
		case <-s.stopCh:
			return
		}
	}
}

func (s *Server) broadcastMessage(msg map[string]interface{}) {
	data, _ := json.Marshal(msg)
	select {
	case s.broadcast <- data:
	default:
		// Channel full, skip
	}
}

func (s *Server) sendWSMessage(conn *websocket.Conn, msg map[string]interface{}) {
	data, _ := json.Marshal(msg)
	//nolint:errcheck
	conn.WriteMessage(websocket.TextMessage, data)
}

func (s *Server) handleShutdown(sigCh chan os.Signal) {
	<-sigCh
	log.Println("Shutting down daemon...")
	//nolint:errcheck
	s.Stop()
}

// Helper functions
func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	//nolint:errcheck
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	//nolint:errcheck
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Scheduler handlers
func (s *Server) schedulerHandler(w http.ResponseWriter, r *http.Request) {
	sch := scheduler.NewScheduler(s.cfg)
	if err := sch.LoadTasks(); err != nil {
		sendError(w, "Failed to load tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]interface{}{
		"tasks": sch.GetTasks(),
		"stats": sch.GetStats(),
	})
}

func (s *Server) schedulerCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Schedule string `json:"schedule"`
		Enabled  bool   `json:"enabled"`
		Tool     string `json:"tool"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sch := scheduler.NewScheduler(s.cfg)
	if err := sch.LoadTasks(); err != nil {
		sendError(w, "Failed to load tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	task := &scheduler.ScheduledTask{
		ID:       req.ID,
		Type:     scheduler.TaskType(req.Type),
		Schedule: req.Schedule,
		Enabled:  req.Enabled,
		Tool:     req.Tool,
	}

	if err := sch.AddTask(task); err != nil {
		sendError(w, "Failed to add task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save to config
	s.cfg.Scheduler.Tasks = sch.ToConfig()
	if err := config.Save(s.cfg, config.GetDefaultConfigPath()); err != nil {
		sendError(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, task)
}

func (s *Server) schedulerUpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Schedule string `json:"schedule"`
		Enabled  bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sch := scheduler.NewScheduler(s.cfg)
	if err := sch.LoadTasks(); err != nil {
		sendError(w, "Failed to load tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := sch.UpdateTask(id, req.Enabled, req.Schedule); err != nil {
		sendError(w, "Failed to update task: "+err.Error(), http.StatusNotFound)
		return
	}

	// Save to config
	s.cfg.Scheduler.Tasks = sch.ToConfig()
	if err := config.Save(s.cfg, config.GetDefaultConfigPath()); err != nil {
		sendError(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	task, _ := sch.GetTask(id)
	sendJSON(w, task)
}

func (s *Server) schedulerDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	sch := scheduler.NewScheduler(s.cfg)
	if err := sch.LoadTasks(); err != nil {
		sendError(w, "Failed to load tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := sch.DeleteTask(id); err != nil {
		sendError(w, "Failed to delete task: "+err.Error(), http.StatusNotFound)
		return
	}

	// Save to config
	s.cfg.Scheduler.Tasks = sch.ToConfig()
	if err := config.Save(s.cfg, config.GetDefaultConfigPath()); err != nil {
		sendError(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]string{"status": "deleted"})
}

func (s *Server) schedulerRunHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	sch := scheduler.NewScheduler(s.cfg)
	if err := sch.LoadTasks(); err != nil {
		sendError(w, "Failed to load tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	task, ok := sch.GetTask(id)
	if !ok {
		sendError(w, "Task not found: "+id, http.StatusNotFound)
		return
	}

	// Run the task
	go sch.RunTask(task)

	sendJSON(w, map[string]interface{}{
		"status": "running",
		"task":   task.ID,
	})
}

// Credentials handlers
func (s *Server) credentialsHandler(w http.ResponseWriter, r *http.Request) {
	store, err := credentials.NewCredentialsStore()
	if err != nil {
		sendError(w, "Failed to create credentials store: "+err.Error(), http.StatusInternalServerError)
		return
	}

	creds, err := store.List()
	if err != nil {
		sendError(w, "Failed to list credentials: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]interface{}{
		"credentials": creds,
		"count":       len(creds),
	})
}

func (s *Server) credentialsSetHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model    string `json:"model"`
		Key      string `json:"key"`
		Value    string `json:"value,omitempty"`
		EnvVar   string `json:"env_var,omitempty"`
		Provider string `json:"provider"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Model == "" || req.Key == "" {
		sendError(w, "Model and key are required", http.StatusBadRequest)
		return
	}

	store, err := credentials.NewCredentialsStore()
	if err != nil {
		sendError(w, "Failed to create credentials store: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if req.EnvVar != "" {
		// Set environment variable reference
		if err := store.SetFromEnv(req.Model, req.Key, req.EnvVar, req.Provider); err != nil {
			sendError(w, "Failed to set credential: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if req.Value != "" {
		// Set value directly (stored securely)
		if err := store.Set(req.Model, req.Key, req.Value, req.Provider); err != nil {
			sendError(w, "Failed to set credential: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		sendError(w, "Either value or env_var is required", http.StatusBadRequest)
		return
	}

	s.broadcastMessage(map[string]interface{}{
		"type":   "credential_set",
		"model":  req.Model,
		"key":    req.Key,
	})

	sendJSON(w, map[string]interface{}{
		"status":  "set",
		"model":   req.Model,
		"key":     req.Key,
		"provider": req.Provider,
	})
}

func (s *Server) credentialsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	model := vars["model"]
	key := vars["key"]

	store, err := credentials.NewCredentialsStore()
	if err != nil {
		sendError(w, "Failed to create credentials store: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := store.Delete(model, key); err != nil {
		sendError(w, "Failed to delete credential: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.broadcastMessage(map[string]interface{}{
		"type":  "credential_deleted",
		"model": model,
		"key":   key,
	})

	sendJSON(w, map[string]interface{}{
		"status": "deleted",
		"model":  model,
		"key":    key,
	})
}

func (s *Server) modelCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	model := vars["model"]

	store, err := credentials.NewCredentialsStore()
	if err != nil {
		sendError(w, "Failed to create credentials store: "+err.Error(), http.StatusInternalServerError)
		return
	}

	creds, err := store.GetForModel(model)
	if err != nil {
		sendError(w, "Failed to get credentials: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check which credentials are actually set
	results := make([]map[string]interface{}, 0, len(creds))
	for _, c := range creds {
		_, err := store.Get(model, c.Key)
		envVar, isEnv, _ := store.GetEnvVar(model, c.Key)

		results = append(results, map[string]interface{}{
			"model":   c.Model,
			"key":     c.Key,
			"source":  c.Source,
			"provider": c.Provider,
			"set":     err == nil || isEnv,
			"env_var": envVar,
		})
	}

	sendJSON(w, map[string]interface{}{
		"model":       model,
		"credentials": results,
	})
}
