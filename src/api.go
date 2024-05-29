package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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
	router.HandleFunc("/account", makeHttpHandleFunc(s.HandleAccount))
	router.HandleFunc("/account/{id}", makeHttpHandleFunc(s.HandleAccountWithId))
	router.HandleFunc("/transfer", makeHttpHandleFunc(s.handleTransfer))

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

	account := NewAccount(accountRequest.FirstName, accountRequest.LastName)

	err := s.storage.CreateAccount(account)
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

func (s *APIServer) handleListAccounts(w http.ResponseWriter, r *http.Request) error {
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
