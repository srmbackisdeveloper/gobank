package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccount(int) (*Account, error)
	GetAccounts() ([]*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=postgres password=admin1618 sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Init() error {
	return s.CreateAccountTable()
}

func (s *PostgresStore) CreateAccountTable() error {
	query := `CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    first_name varchar(50),
    last_name varchar(50),
    number serial,
    balance serial,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`
	_, err := s.db.Exec(query)

	return err
}

// ---------------------------------------------------------------

func (s *PostgresStore) CreateAccount(account *Account) error {
	query := `INSERT INTO accounts (first_name, last_name, number, balance, created_at) VALUES ($1, $2, $3, $4, $5);`

	res, err := s.db.Query(query,
		account.FirstName, account.LastName, account.Number, account.Balance, account.CreatedAt)

	if err != nil {
		return err
	}

	fmt.Printf("Insertion successfully done: %+v\n", res)

	return nil
}

func (s *PostgresStore) UpdateAccount(*Account) error {
	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM accounts WHERE id=$1)", id).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("account with id=%d does not exist", id)
	}

	query := "DELETE FROM accounts WHERE id=$1;"

	_, err = s.db.Query(query, id)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) GetAccount(id int) (*Account, error) {
	query := `SELECT * FROM accounts WHERE id=$1;`

	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return ScanIntoAccount(rows)
	}

	return nil, fmt.Errorf("the account (id: %v) is not found", id)
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	query := `SELECT * FROM accounts;`

	rows, err := s.db.Query(query)

	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := ScanIntoAccount(rows)

		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	// fmt.Printf("%+v\n", res)

	return accounts, nil
}

func ScanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)

	err := rows.Scan(
		&account.Id,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.Balance,
		&account.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return account, nil
}
