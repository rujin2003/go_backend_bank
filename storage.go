package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Storage interface for account storage operations.
type Storage interface {
	CheckAuth(string, string) error
	CreateAccount(*account) error
	DeleteAccount(int) error
	UpdateAccount(*account) error
	GetAccountByID(int) (*account, error)
	GetUsers() ([]*account, error)
	Close()
}

// PostgresStorage struct for PostgreSQL storage.
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage initializes a new PostgresStorage instance.

func NewPostgresStorage() (*PostgresStorage, error) {
	connStr := "user=postgres password=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Check if the database exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = 'bank')").Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		// Create the database if it does not exist
		_, err = db.Exec("CREATE DATABASE bank")
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
	}

	// Connect to the newly created or existing database
	db, err = sql.Open("postgres", connStr+" dbname=bank")
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

// Init initializes the database by creating necessary tables.
func (s *PostgresStorage) Init() error {
	_, err := s.db.Exec(`
        CREATE TABLE IF NOT EXISTS accounts (
            id SERIAL PRIMARY KEY,
            email TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL,
            name TEXT,
            number TEXT,
            balance INT
        )
    `)
	return err
}

// CreateAccount inserts a new account into the database.
func (s *PostgresStorage) CreateAccount(a *account) error {
	err := s.db.QueryRow(
		"INSERT INTO accounts (email, password, name, number, balance) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		a.Email, a.Password, a.Name, a.Number, a.Balance,
	).Scan(&a.ID)
	return err
}

// CheckAuth checks if the provided email and password match the stored account.

func (s *PostgresStorage) CheckAuth(email string, password string) error {
	row := s.db.QueryRow("SELECT password FROM accounts WHERE email = $1", email)
	a := &account{}
	err := row.Scan(&a.Password)
	if err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	if err != nil {
		return fmt.Errorf("authentication failed: incorrect password")
	}

	return nil
}

func (s *PostgresStorage) GetUsers() ([]*account, error) {
	rows, err := s.db.Query("SELECT id, name, number, balance FROM accounts") // could be replaced with "SELECT * FROM accounts"

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]*account, 0)
	for rows.Next() {
		a := &account{}
		err := rows.Scan(&a.ID, &a.Name, &a.Number, &a.Balance)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}

	return accounts, nil
}

// DeleteAccount deletes an account from the database by its ID.

func (s *PostgresStorage) DeleteAccount(id int) error {
	_, err := s.db.Exec("DELETE FROM accounts WHERE id = $1", id)
	fmt.Printf("Deleted account with id: %d\n", id)
	return err
}

// UpdateAccount updates an existing account in the database.
func (s *PostgresStorage) UpdateAccount(a *account) error {
	_, err := s.db.Exec("UPDATE accounts SET name = $1, number = $2, balance = $3 WHERE id = $4", a.Name, a.Number, a.Balance, a.ID)
	return err
}

// GetAccountByID retrieves an account from the database by its ID.
func (s *PostgresStorage) GetAccountByID(id int) (*account, error) {
	row := s.db.QueryRow("SELECT id, name, number, balance FROM accounts WHERE id = $1", id)
	a := &account{}
	err := row.Scan(&a.ID, &a.Name, &a.Number, &a.Balance)
	return a, err
}

// Close closes the database connection.
func (s *PostgresStorage) Close() {
	s.db.Close()
}
