package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
)

const (
	size              = 4
	userCookie        = "user_id"
	noUserIDError     = "unable to verify user, no user_id provided"
	verifyError       = "unable to verify user via sign"
	IDError           = "unable to create user_id"
	signCreationError = "unable to sign cookie"
)

type BasicAuth struct {
	SecretKey string
}

func NewBasicAuth(secretKey string) *BasicAuth {
	return &BasicAuth{
		SecretKey: secretKey,
	}
}

func (b *BasicAuth) CreateRandomID() (string, error) {
	sl := make([]byte, size)
	_, err := rand.Read(sl)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sl), nil
}

func (b *BasicAuth) CreateNewSign() (string, error) {
	hash1 := hmac.New(md5.New, []byte(b.SecretKey))
	_, err := hash1.Write([]byte(b.SecretKey))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash1.Sum(nil)), nil
}

func (b *BasicAuth) AuthentificateUser(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := r.Cookie(userCookie)
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, noUserIDError, http.StatusUnauthorized)
			return
		} else {
			hash2 := hmac.New(md5.New, []byte(b.SecretKey))
			_, err = hash2.Write([]byte(b.SecretKey))
			if err != nil {
				http.Error(w, verifyError, http.StatusUnauthorized)
				return
			}
			sign := hash2.Sum(nil)
			cookie, err := hex.DecodeString(userID.Value)
			if err != nil || !hmac.Equal(sign, cookie[size:]) {
				http.Error(w, verifyError, http.StatusUnauthorized)
				return
			}
		}
		h(w, r)
	}
}

func (b *BasicAuth) CheckUserCookies(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := r.Cookie(userCookie)
		if errors.Is(err, http.ErrNoCookie) {
			id, err := b.CreateRandomID()
			if err != nil {
				http.Error(w, IDError, http.StatusUnauthorized)
				return
			}

			sign, err := b.CreateNewSign()
			if err != nil {
				http.Error(w, signCreationError, http.StatusUnauthorized)
				return
			}

			cookieVal := id + sign
			http.SetCookie(w, &http.Cookie{
				Name:  userCookie,
				Value: cookieVal,
			})
		} else {
			hash2 := hmac.New(md5.New, []byte(b.SecretKey))
			hash2.Write([]byte(b.SecretKey))
			sign := hash2.Sum(nil)
			cook, err := hex.DecodeString(userID.Value)

			if err != nil || !hmac.Equal(cook[size:], sign) {
				id, err := b.CreateRandomID()
				if err != nil {
					http.Error(w, IDError, http.StatusUnauthorized)
					return
				}

				newSign, err := b.CreateNewSign()
				if err != nil {
					http.Error(w, signCreationError, http.StatusUnauthorized)
					return
				}

				cookieVal := id + newSign
				http.SetCookie(w, &http.Cookie{
					Name:  userCookie,
					Value: cookieVal,
				})
			}
		}
		h(w, r)
	}
}
