package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"kickoff.com/pkg/discovery"
	"kickoff.com/pkg/discovery/consul"
	"kickoff.com/pkg/models"
)

const serviceName = "user"

type UserService struct {
	users   map[string]models.User
	counter int
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "API handler port")
	flag.Parse()

	log.Printf("Starting user service on port %d", port)

	// Crear conexión con Consul
	consulAddr := os.Getenv("CONSUL_ADDRESS")
	if consulAddr == "" {
		log.Fatal("CONSUL_ADDRESS environment variable is required")
	}
	registry, err := consul.NewRegistry(consulAddr)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)

	// Registrar servicio en Consul con el nombre del contenedor
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("user-service:%d", port)); err != nil {
		panic(err)
	}

	// Goroutine para reportar estado de salud cada segundo
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Asegurar que se desregistre al terminar
	defer registry.Deregister(ctx, instanceID, serviceName)

	// Inicializar servicio
	userService := &UserService{
		users:   make(map[string]models.User),
		counter: 1,
	}

	// Crear algunos usuarios de ejemplo
	userService.createSampleUsers()

	// === ENDPOINTS EXISTENTES (mantener funcionando) ===
	http.HandleFunc("/health", userService.healthHandler)
	http.HandleFunc("/users", userService.usersHandlerOld) // Endpoint original

	// === NUEVOS ENDPOINTS CON MODELOS ROBUSTOS ===
	http.HandleFunc("/v2/users", userService.usersHandlerNew)           // CRUD completo
	http.HandleFunc("/v2/users/", userService.userByIDHandler)          // Usuario por ID
	http.HandleFunc("/v2/users/search/", userService.userSearchHandler) // Búsqueda por username/email

	log.Printf("User service listening on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}

func (us *UserService) createSampleUsers() {
	sampleUsers := []models.CreateUserRequest{
		{Username: "user1", Email: "user1@kickoff.com", FullName: "John Doe"},
		{Username: "user2", Email: "user2@kickoff.com", FullName: "Jane Smith"},
		{Username: "nflFan123", Email: "nfl.fan@gmail.com", FullName: "Mike Johnson"},
	}

	for _, req := range sampleUsers {
		us.createUser(req)
	}
}

func (us *UserService) createUser(req models.CreateUserRequest) models.User {
	user := models.User{
		ID:        fmt.Sprintf("user_%d", us.counter),
		Username:  req.Username,
		Email:     req.Email,
		FullName:  req.FullName,
		CreatedAt: time.Now(),
		Active:    true,
	}
	us.counter++
	us.users[user.ID] = user
	return user
}

func (us *UserService) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User service is healthy"))
}

// === ENDPOINT ORIGINAL (NO TOCAR) ===
func (us *UserService) usersHandlerOld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "User service is running", "service": "user"}`))
}

// === NUEVOS ENDPOINTS CON MODELOS ROBUSTOS ===
func (us *UserService) usersHandlerNew(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// Obtener todos los usuarios
		var users []models.User
		for _, user := range us.users {
			users = append(users, user)
		}

		response := map[string]interface{}{
			"users": users,
			"total": len(users),
		}
		json.NewEncoder(w).Encode(response)

	case "POST":
		// Crear nuevo usuario
		var req models.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validar campos requeridos
		if req.Username == "" || req.Email == "" || req.FullName == "" {
			http.Error(w, "Missing required fields: username, email, fullName", http.StatusBadRequest)
			return
		}

		// Verificar que username no exista
		for _, user := range us.users {
			if user.Username == req.Username {
				http.Error(w, "Username already exists", http.StatusConflict)
				return
			}
		}

		// Verificar que email no exista
		for _, user := range us.users {
			if user.Email == req.Email {
				http.Error(w, "Email already exists", http.StatusConflict)
				return
			}
		}

		// Crear usuario
		user := us.createUser(req)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (us *UserService) userByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extraer user ID de la URL (/v2/users/{userID})
	userID := strings.TrimPrefix(r.URL.Path, "/v2/users/")

	switch r.Method {
	case "GET":
		user, exists := us.users[userID]
		if !exists {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(user)

	case "PUT":
		// Actualizar usuario
		user, exists := us.users[userID]
		if !exists {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		var req models.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Actualizar campos que no estén vacíos
		if req.Username != "" {
			user.Username = req.Username
		}
		if req.Email != "" {
			user.Email = req.Email
		}
		if req.FullName != "" {
			user.FullName = req.FullName
		}
		if req.Active != nil {
			user.Active = *req.Active
		}

		us.users[userID] = user
		json.NewEncoder(w).Encode(user)

	case "DELETE":
		// Desactivar usuario (soft delete)
		user, exists := us.users[userID]
		if !exists {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		user.Active = false
		us.users[userID] = user

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (us *UserService) userSearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraer término de búsqueda de la URL (/v2/users/search/{term})
	searchTerm := strings.TrimPrefix(r.URL.Path, "/v2/users/search/")
	searchTerm = strings.ToLower(searchTerm)

	if searchTerm == "" {
		http.Error(w, "Search term is required", http.StatusBadRequest)
		return
	}

	var matchingUsers []models.User
	for _, user := range us.users {
		if strings.Contains(strings.ToLower(user.Username), searchTerm) ||
			strings.Contains(strings.ToLower(user.Email), searchTerm) ||
			strings.Contains(strings.ToLower(user.FullName), searchTerm) {
			matchingUsers = append(matchingUsers, user)
		}
	}

	response := map[string]interface{}{
		"users":      matchingUsers,
		"total":      len(matchingUsers),
		"searchTerm": searchTerm,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
