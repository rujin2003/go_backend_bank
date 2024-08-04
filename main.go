package main

import (
	"encoding/json"
	"fmt"

	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Apiserver struct holds the server's address and a storage interface.
type Apiserver struct {
	listenAddress string
	store         Storage
}

// NewApiServer initializes a new instance of Apiserver with the provided address.
func NewApiServer(listenAddress string) *Apiserver {
	return &Apiserver{listenAddress: listenAddress}
}

// Run starts the API server and sets up the routes.
func (s *Apiserver) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/account", makeHandler(s.handleAccount)).Methods("GET", "POST")

	router.Handle("/login", makeHandler(s.handleLogin)).Methods("POST")

	router.HandleFunc("/account/users", makeHandler(s.handleGetUsers)).Methods("GET")
	router.HandleFunc("/account/{id}", ProtectedHandler(s.handleGetAccountById)).Methods("GET", "DELETE")
	router.HandleFunc("/account/create", makeHandler(s.handleCreateAccount)).Methods("POST")

	router.HandleFunc("/transfer", makeHandler(s.handleTransfer)).Methods("POST")

	http.ListenAndServe(s.listenAddress, router)
}

func (s *Apiserver) handleLogin(w http.ResponseWriter, r *http.Request) error {

	loginRequest := LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		return err
	}

	err := s.store.CheckAuth(loginRequest.Email, loginRequest.Password)

	if err != nil {

		return writeJSON(w, http.StatusUnauthorized, ApiError{Error: err.Error()})
	} else {
		tokenString, JWTerr := CreateToken(loginRequest.Email)
		if JWTerr != nil {
			fmt.Print("No username found")
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, tokenString)
	}

	return writeJSON(w, http.StatusOK, map[string]string{"message": "login successful"})
}

// handleAccount handles requests to the /account endpoint based on the HTTP method.
func (s *Apiserver) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccountById(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("unsupported method")
}

// handleGetAccount handles GET requests to retrieve account information.
func (s *Apiserver) handleGetAccountById(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		vars := mux.Vars(r)["id"]
		id, err := strconv.Atoi(vars)
		if err != nil {
			return err // return error if conversion fails
		}
		users, err := s.store.GetAccountByID(id)
		if err != nil {
			return err
		}

		return writeJSON(w, http.StatusOK, users)
	} else {
		s.handleDeleteAccount(w, r)
		return nil
	}
}

// get all users
func (s *Apiserver) handleGetUsers(w http.ResponseWriter, r *http.Request) error {
	// Retrieve all users from the database
	users, err := s.store.GetUsers()
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, users)

}

// handleCreateAccount handles POST requests to create a new account.
func (s *Apiserver) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	CreateAccountReq := CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(&CreateAccountReq); err != nil {
		return err
	}

	acc, err := NewAccount(CreateAccountReq.Email, CreateAccountReq.Password, CreateAccountReq.Name, CreateAccountReq.Number, CreateAccountReq.Balance)
	if err != nil {
		return err
	}

	if err := s.store.CreateAccount(acc); err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, CreateAccountReq)
}

// handleDeleteAccount handles DELETE requests to delete an account.
func (s *Apiserver) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)["id"]
	id, err := strconv.Atoi(vars)
	if err != nil {
		return err
	}
	users := s.store.DeleteAccount(id)

	return writeJSON(w, http.StatusOK, users)

}

// handleTransfer handles POST requests to transfer funds between accounts.
func (s *Apiserver) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	// Implement funds transfer logic here
	return nil
}

// writeJSON writes a JSON response to the ResponseWriter.
func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// apiFunc type is a function that handles an HTTP request and returns an error.
type apiFunc func(w http.ResponseWriter, r *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

// makeHandler wraps an apiFunc and converts it to an http.HandlerFunc.
func makeHandler(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			writeJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}

}

func ProtectedHandler(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Missing authorization header")
			return
		}
		tokenString := authHeader[len("Bearer "):]

		err := verifyToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Invalid token: %v", err)
			return
		}

		if err := fn(w, r); err != nil {
			writeJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

// main function initializes and runs the API server.

func main() {

	store, err := NewPostgresStorage()

	if err != nil {
		fmt.Println("Failed to initialize storage:", err)
		return
	}
	defer store.Close()

	// Initialize the database (create tables)
	if err := store.Init(); err != nil {
		fmt.Println("Failed to initialize database:", err)
		return
	}

	server := NewApiServer(":3000")
	server.store = store
	server.Run()
}
