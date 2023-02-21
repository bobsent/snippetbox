package models

import "errors"

var (
	ErrNoRecord = errors.New("models: no matching records found")

	// Add a new ErrInvalidCredentials error. We'll use this later if a user
	// tries to login with an incorrect email address or passsword.
	ErrInvalidCredentials = errors.New("models: no matching record found")

	// Add a new ErrDuplicateEmail error. We'll use this later if a user
	// tries to login with an incorrect email address or password
	ErrDuplicateEmail = errors.New("models: duplicate email")
)
