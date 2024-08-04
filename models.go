// CreateAccountRequest struct represents a request to create a new account.
package main

import (
	"golang.org/x/crypto/bcrypt"
)

type CreateAccountRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Number   string `json:"number"`
	Balance  int    `json:"balance"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// account struct represents an account entity.
type account struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Number   string `json:"number"`
	Balance  int    `json:"balance"`
}

// NewAccount creates a new account instance.
func NewAccount(email string, password string, name, number string, balance int) (*account, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &account{
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
		Number:   number,
		Balance:  balance,
	}, nil
}
