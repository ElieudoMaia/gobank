package main

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountById(int) (*Account, error)
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	connStr := "user=gobank dbname=gobank password=gobank sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) InitDB() error {
	err := s.db.Ping()
	if err != nil {
		return err
	}

	err = s.RunMigrations()
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) RunMigrations() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id SERIAL PRIMARY KEY,
			first_name TEXT,
			last_name TEXT,
			number BIGINT,
			balance FLOAT
		)
	`)
	if err != nil {
		return err
	}
	return err
}

func (s *PostgresStorage) CreateAccount(account *Account) error {
	_, err := s.db.Exec(`
	INSERT INTO accounts (
		first_name,
		last_name,
		number,
		balance
	)
	VALUES ($1, $2, $3, $4)`,
		account.FirstName, account.LastName, account.Number, account.Balance,
	)

	return err
}

func (s *PostgresStorage) DeleteAccount(id int) error {
	_, err := s.db.Exec("DELETE FROM accounts WHERE id = $1", id)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) UpdateAccount(account *Account) error {
	_, err := s.db.Exec("UPDATE accounts SET first_name = $1, last_name = $2 WHERE id = $3", account.FirstName, account.LastName, account.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) GetAccountById(id int) (*Account, error) {
	row := s.db.QueryRow("SELECT id, first_name, last_name FROM accounts WHERE id = $1", id)
	account := &Account{}
	err := row.Scan(&account.ID, &account.FirstName, &account.LastName)
	if err != nil {
		return nil, err
	}
	return account, nil
}
