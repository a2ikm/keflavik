package app

import (
	"github.com/a2ikm/keflavik/model"
)

type App struct {
	Queries *model.Queries
}

type Error struct {
	Code string
}

func (e *Error) Error() string {
	return e.Code
}

var (
	ErrUnauthorized = &Error{"unauthorized"}
)
