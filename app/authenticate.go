package app

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"

	"github.com/a2ikm/keflavik/model"
	"github.com/jackc/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type AuthenticateResult struct {
	AccessToken string
}

func (a *App) Authenticate(ctx context.Context, name string, password string) (AuthenticateResult, error) {
	user, err := a.Queries.GetUserByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return AuthenticateResult{}, ErrUnauthorized
		} else {
			return AuthenticateResult{}, err
		}
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return AuthenticateResult{}, ErrUnauthorized
	}

	var session model.Session
	for {
		token, err := generateRandomString(64)
		if err != nil {
			continue
		}

		params := model.CreateSessionParams{
			UserID:      user.ID,
			AccessToken: token,
		}
		session, err = a.Queries.CreateSession(ctx, params)
		if err != nil {
			if isUniquenessViolation(err) {
				continue
			}
			return AuthenticateResult{}, err
		}

		break
	}

	return AuthenticateResult{
		AccessToken: session.AccessToken,
	}, nil
}

func isUniquenessViolation(err error) bool {
	pgerr, ok := err.(*pgconn.PgError)
	if !ok {
		return false
	}

	return pgerr.Code == "23505"
}

func generateRandomString(digit uint32) (string, error) {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, digit)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	var result string
	for _, v := range b {
		result += string(letters[int(v)%len(letters)])
	}
	return result, nil
}
