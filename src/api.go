package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	storage    Storage
}

func NewAPIServer(listenAddr string, storage Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		storage:    storage,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/account", withAuthentication(makeHttpHandleFunc(s.HandleAccount)))
	router.HandleFunc("/account/{id}", withAuthentication(makeHttpHandleFunc(s.HandleAccountWithId)))
	router.HandleFunc("/transfer", withAuthentication(makeHttpHandleFunc(s.handleTransfer)))
	router.HandleFunc("/signin", makeHttpHandleFunc(s.handleSignIn))

	log.Println("Starting server on", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) HandleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodPost {
		return s.handleCreateAccount(w, r)
	} else if r.Method == http.MethodGet {
		return s.handleListAccounts(w, r)
	}

	return fmt.Errorf("method not allowed")
}

func (s *APIServer) HandleAccountWithId(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return s.handleGetAccount(w, r)
	} else if r.Method == http.MethodDelete {
		return s.handleDeleteAccount(w, r)
	} else if r.Method == http.MethodPut {
		return s.handleUpdateAccount(w, r)
	}

	return fmt.Errorf("method not allowed")
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	idint, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid id")
	}

	account, err := s.storage.GetAccountById(idint)
	if err != nil {
		return err
	}

	account.ID = idint
	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	accountRequest := &CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(accountRequest); err != nil {
		return err
	}

	hashedPassword, err := HashPassword(accountRequest.Password)
	if err != nil {
		return err
	}

	account := NewAccount(accountRequest.FirstName, accountRequest.LastName, string(hashedPassword))

	err = s.storage.CreateAccount(account)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	idint, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid id")
	}

	err = s.storage.DeleteAccount(idint)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, nil)
}

func (s *APIServer) handleListAccounts(w http.ResponseWriter, _ *http.Request) error {
	accounts, err := s.storage.ListAccounts()
	if err != nil {
		return errors.New("erro trying list accounts")
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleUpdateAccount(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	idint, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid id")
	}

	savedAccount, err := s.storage.GetAccountById(idint)
	if err != nil {
		return err
	}
	if savedAccount == nil {
		return fmt.Errorf("account not found")
	}

	accountRequest := &UpdateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(accountRequest); err != nil {
		return err
	}

	savedAccount.FirstName = accountRequest.FirstName
	savedAccount.LastName = accountRequest.LastName

	err = s.storage.UpdateAccount(savedAccount)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, savedAccount)
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPatch {
		return errors.New("method not allowed")
	}

	transferRequest := &TransferRequest{}
	if err := json.NewDecoder(r.Body).Decode(transferRequest); err != nil {
		return err
	}

	accountFrom, err := s.storage.GetAccountById(transferRequest.FromAccountID)
	if err != nil {
		return errors.New("account from not found")
	}

	accountTo, err := s.storage.GetAccountById(transferRequest.ToAccountID)
	if err != nil {
		return errors.New("account to not found")
	}

	accountFrom.Balance = accountFrom.Balance - transferRequest.Amount
	accountTo.Balance = accountTo.Balance + transferRequest.Amount

	fmt.Print(accountFrom)
	fmt.Print(accountTo)

	err = s.storage.SaveBalance(accountFrom, accountTo)

	return err
}

func (s *APIServer) handleSignIn(w http.ResponseWriter, r *http.Request) error {
	signInRequest := &SignInRequest{}

	if err := json.NewDecoder(r.Body).Decode(signInRequest); err != nil {
		return err
	}

	account, err := s.storage.GetAccountByNumber(signInRequest.AccountNumber)
	if err != nil {
		return err
	}

	if account == nil {
		return errors.New("account not found")
	}

	err = CompareHashAndPassword(account.Password, signInRequest.Password)
	if err != nil {
		return errors.New("invalid password")
	}

	token, err := GenerateJWTToken(account)
	if err != nil {
		return err
	}

	response := struct {
		AccessToken string `json:"access_token"`
	}{AccessToken: token}

	return WriteJSON(w, http.StatusOK, response)
}

// helper to write http json reponse
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
	return nil
}

// Type of http custom handlers
type Controller func(w http.ResponseWriter, r *http.Request) error

type APIError struct {
	Error string `json:"error"`
}

// adapter for http custom handlers
func makeHttpHandleFunc(c Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := c(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		}
	}
}

var tokenRegex = regexp.MustCompile(`[^\s]+`)

func withAuthentication(httpHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header["Authorization"]
		if header == nil {
			response := struct {
				Message string `json:"message"`
			}{Message: "missing token"}
			WriteJSON(w, http.StatusUnauthorized, response)
			return
		}
		headerValue := tokenRegex.FindAllString(header[0], -1)

		token := headerValue[1]

		accountID, err := VerifyToken(token)
		if err != nil {
			response := struct {
				Message string `json:"message"`
			}{Message: "invalid token"}
			WriteJSON(w, http.StatusUnauthorized, response)
			return
		}

		log.Print("Request From AccountID = ", accountID)

		httpHandler(w, r)
	}
}
