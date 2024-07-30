package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const size = 4

type BasicAuth struct {
	SecretKey string
}

func NewBasicAuth(secretKey string) *BasicAuth {
	return &BasicAuth{
		SecretKey: secretKey,
	}
}

func (b *BasicAuth) CreateRandomID() string {
	sl := make([]byte, size)
	rand.Read(sl)
	return hex.EncodeToString(sl)
}

func (b *BasicAuth) CreateNewSign() string {
	hash1 := hmac.New(md5.New, []byte(b.SecretKey))
	hash1.Write([]byte(b.SecretKey))
	sign := hash1.Sum(nil)
	return hex.EncodeToString(sign)
}

func (b *BasicAuth) Authenticate(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := r.Cookie("user_id")
		if err != nil {
			http.Error(w, "unable to verify user, no user_id provided", http.StatusUnauthorized)
			return
		} else {
			hash2 := hmac.New(md5.New, []byte(b.SecretKey))
			hash2.Write([]byte(b.SecretKey))
			sign := hash2.Sum(nil)
			cook, _ := hex.DecodeString(userID.Value)
			if !hmac.Equal(sign, cook[size:]) {
				http.Error(w, "unable to verify user via sign", http.StatusUnauthorized)
				return
			}
		}
		h(w, r)
	}
}

func (b *BasicAuth) CheckCookies(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := r.Cookie("user_id")
		if err != nil {
			cook := b.CreateRandomID() + b.CreateNewSign()
			http.SetCookie(w, &http.Cookie{
				Name:  "user_id",
				Value: cook,
			})
		} else {
			hash2 := hmac.New(md5.New, []byte(b.SecretKey))
			hash2.Write([]byte(b.SecretKey))
			sign := hash2.Sum(nil)
			cook, _ := hex.DecodeString(userID.Value)
			if !hmac.Equal(cook[size:], sign) {
				http.SetCookie(w, &http.Cookie{
					Name:  "user_id",
					Value: b.CreateRandomID() + b.CreateNewSign(),
				})
			}
		}
		h(w, r)
	}
}
