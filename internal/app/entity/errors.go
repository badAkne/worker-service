package entity

import "errors"

// Доменные ошибки приложения.
// Используются в handlers и modules, маппятся на HTTP коды в respondent.
var (
	// ErrNotFound — ресурс не найден.
	ErrNotFound = errors.New("entity: not found")

	// ErrAlreadyExists — ресурс уже существует.
	ErrAlreadyExists = errors.New("entity: already exists")

	// ErrInvalidInput — некорректные входные данные.
	ErrInvalidInput = errors.New("entity: invalid input")

	// ErrForbidden — доступ запрещён.
	ErrForbidden = errors.New("entity: forbidden")

	// ErrUnauthorized — требуется авторизация.
	ErrUnauthorized = errors.New("entity: unauthorized")
)
