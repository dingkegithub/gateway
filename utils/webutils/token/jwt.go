package token

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

var (
	identify = []byte("user-identify-key")
)

type UserPayload struct {
	Uid int64 `json:"uid"`
	Ip string `json:"ip"`
	DeviceId string `json:"device_id"`
}

type UserClaim struct {
	*UserPayload
	*jwt.StandardClaims
}


func Encode(m *UserPayload) (string, error) {
	claim := &UserClaim{
		UserPayload:    m,
		StandardClaims: &jwt.StandardClaims{
			Audience:  "",
			ExpiresAt: time.Now().Unix(),
			Id:        "",
			IssuedAt:  0,
			Issuer:    "",
			NotBefore: 0,
			Subject:   "",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(identify)
}

func Decode(tokenStr string) (payload *UserPayload, expire bool, err error)  {

	claim := &UserClaim{}
	tk, err := jwt.ParseWithClaims(tokenStr, claim, func(token *jwt.Token) (i interface{}, e error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ParseError
		}
		return identify, nil
	})

	if err != nil {
		if v, ok := err.(*jwt.ValidationError); ok {
			if v.Errors & jwt.ValidationErrorExpired == 0 {
				return claim.UserPayload, true, nil
			}
		}
		return nil, false, FormatError
	}

	if tk == nil {
		return nil, false, FormatError
	}

	if tk.Valid {
		return claim.UserPayload, false, nil
	}

	return nil, false, ValidError
}
