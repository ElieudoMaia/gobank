package main

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountById(int) (*Account, error)
	ListAccounts() ([]*Account, error)
	SaveBalance(accountFrom *Account, accountTo *Account) error
	GetAccountByNumber(int) (*Account, error)
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

	err = s.runMigrations()
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) runMigrations() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id SERIAL PRIMARY KEY,
			first_name varchar(255),
			last_name varchar(255),
			password varchar(255),
			number BIGINT,
			balance FLOAT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	account := NewAccount("Elieudo", "Maia", "123456")
	account.Number = 7844426807732681064

	row := s.db.QueryRow("SELECT * FROM accounts where number = $1 LIMIT 1", account.Number)
	err = row.Scan(&account.ID, &account.FirstName, &account.LastName, &account.Password, &account.Number, &account.Balance, &account.CreatedAt)
	if err == nil {
		return nil
	}

	hashedPassword, err := HashPassword(account.Password)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`INSERT INTO accounts (first_name, last_name, password, number, balance) VALUES ($1, $2, $3, $4, $5)`,
		account.FirstName,
		account.LastName,
		hashedPassword,
		account.Number,
		account.Balance,
	)

	return err
}

func (s *PostgresStorage) CreateAccount(account *Account) error {
	_, err := s.db.Exec(`
	INSERT INTO accounts (
		first_name,
		last_name,
		password,
		number,
		balance
	)
	VALUES ($1, $2, $3, $4)`,
		account.FirstName, account.LastName, account.Password, account.Number, account.Balance,
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

func (s *PostgresStorage) ListAccounts() ([]*Account, error) {
	rows, err := s.db.Query("SELECT id, first_name, last_name, number, balance FROM accounts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []*Account{}
	for rows.Next() {
		account := &Account{}
		err := rows.Scan(&account.ID, &account.FirstName, &account.LastName, &account.Number, &account.Balance)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (s *PostgresStorage) SaveBalance(accountFrom *Account, accountTo *Account) error {
	_, err := s.db.Query("UPDATE accounts SET balance = $1 WHERE id = $2", accountFrom.Balance, accountFrom.ID)
	if err != nil {
		return err
	}

	_, err = s.db.Query("UPDATE accounts SET balance = $1 WHERE id = $2", accountTo.Balance, accountTo.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) GetAccountByNumber(accountnumber int) (*Account, error) {
	row := s.db.QueryRow("SELECT * FROM accounts WHERE number = $1", accountnumber)
	account := &Account{}
	err := row.Scan(&account.ID, &account.FirstName, &account.LastName, &account.Password, &account.Number, &account.Balance, &account.CreatedAt)
	if err != nil {
		return nil, errors.New("account not found")
	}

	return account, nil
}
