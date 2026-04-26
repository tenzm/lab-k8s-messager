package domain

import "errors"

var (
	ErrEmptyName = errors.New("name cannot be empty")
	ErrNotFound  = errors.New("user not found")
)
