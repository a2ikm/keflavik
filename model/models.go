// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0

package model

import ()

type User struct {
	ID           int32
	Name         string
	PasswordHash string
}