package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func createJWT(account *Account) (string, error) {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file")
		return "", err
	}

	jwt_ := os.Getenv("JWT_SECRET")

	if jwt_ == "" {
		fmt.Println("The environment variable 'JWT_SECRET' is not set.")
	}

	claims := &jwt.MapClaims{
		"ExpiresAt":     jwt.NewNumericDate(time.Unix(1516239022, 0)),
		"accountNumber": account.Number,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwt_))
}

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFeHBpcmVzQXQiOjE1MTYyMzkwMjIsImFjY291bnROdW1iZXIiOjk3MTM0fQ.MDzBAVLj1BOAnDnYhjYk_70dY3bpIAzzQVKL-kkwv7k

func withJWTAuth(handler http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Calling JWT Auth MW\n")

		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)
		if err != nil || !token.Valid {
			WriteJSON(w, http.StatusForbidden, APIError{Error: "access denied"})
			return
		}

		idString := mux.Vars(r)["id"]
		id, err := strconv.Atoi(idString)
		if err != nil {
			WriteJSON(w, http.StatusForbidden, APIError{Error: "access denied"})
			return
		}

		account, err := s.GetAccount(id)
		if err != nil {
			WriteJSON(w, http.StatusForbidden, APIError{Error: "access denied"})
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		// panic(reflect.TypeOf(claims["accountNumber"]))

		if account.Number != int64(claims["accountNumber"].(float64)) {
			WriteJSON(w, http.StatusForbidden, APIError{Error: "access denied"})
			return
		}

		handler(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file")
		return nil, err
	}

	jwt_ := os.Getenv("JWT_SECRET")

	if jwt_ == "" {
		fmt.Println("The environment variable 'JWT_SECRET' is not set.")
	}

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(jwt_), nil
	})
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			err := WriteJSON(w, http.StatusBadRequest, APIError{Error: "something went wrong: " + err.Error()})
			if err != nil {
				return
			}
		}
	}
}

type APIError struct {
	Error string `json:"error"`
}

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/login", makeHttpHandleFunc(s.handleLogin))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHttpHandleFunc(s.handleAccountWithId), s.store))
	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))

	router.HandleFunc("/transaction", makeHttpHandleFunc(s.handleTransaction))

	log.Println("API server running on port: ", s.listenAddr)

	err := http.ListenAndServe(s.listenAddr, router)
	if err != nil {
		log.Fatal("server killed")
	}

}

// -------- \\

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	var loginReq LoginRequest

	if r.Method != "POST" {
		return fmt.Errorf("method is not allowed %v", r.Method)
	}

	/*
		tokenString, err := createJWT(account)
		if err != nil {
			return err
		}

		fmt.Printf("JWT: %s", tokenString)
	*/

	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, loginReq)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccounts(w, r)
	}

	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("method is not allowed %v", r.Method)
}

func (s *APIServer) handleAccountWithId(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return fmt.Errorf("method is not allowed %v", r.Method)
}

// -------- \\
// void foo() { cout << "hello"; }

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	idString := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idString)
	if err != nil {
		return fmt.Errorf("invalid id=%v is provided", idString)
	}

	account, err := s.store.GetAccount(id)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleGetAccounts(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFeHBpcmVzQXQiOjE1MTYyMzkwMjIsImFjY291bnROdW1iZXIiOjIwOTkxfQ.jCxawtv3090pO8-qNUYW3LG1cXaQA1j_2YSJzbyyelI

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {

	req := new(CreateAccountRequest)

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	account := NewAccount(req.FirstName, req.LastName)

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	tokenString, err := createJWT(account)
	if err != nil {
		return err
	}

	fmt.Printf("JWT: %s", tokenString)
	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	idString := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idString)
	if err != nil {
		return fmt.Errorf("invalid id=%v is provided", idString)
	}

	if err = s.store.DeleteAccount(id); err != nil {
		return err
	}

	// return fmt.Errorf("account with id=%v does not exist", idString)

	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
}

func (s *APIServer) handleTransaction(w http.ResponseWriter, r *http.Request) error {
	transactionReq := new(TransactionRequest)

	if err := json.NewDecoder(r.Body).Decode(transactionReq); err != nil {
		return err
	}

	defer r.Body.Close()

	return WriteJSON(w, http.StatusOK, transactionReq)
}
