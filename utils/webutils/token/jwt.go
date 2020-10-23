package token

import (
	"encoding/json"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	// ToDo:
	// 这个是 jwt secret-salt，不同服务提供不同
	// 不应该直接写在库里
	identify = []byte("gwgwtsignedstr")
)

type UserPayload map[string]interface{}

func (up UserPayload) String() string {
	s, err := json.Marshal(up)
	if err != nil {
		return err.Error()
	}

	return string(s)
}

type UserClaim struct {
	UserPayload
	*jwt.StandardClaims
}

func Encode(m UserPayload, expirs time.Duration) (string, error) {
	claim := &UserClaim{
		UserPayload: m,
		StandardClaims: &jwt.StandardClaims{
			Audience:  "",
			ExpiresAt: time.Now().Add(expirs).Unix(),
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

func Decode(tokenStr string) (payload UserPayload, expire bool, err error) {

	claim := &UserClaim{}
	tk, err := jwt.ParseWithClaims(tokenStr, claim, func(token *jwt.Token) (i interface{}, e error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ParseError
		}
		return identify, nil
	})

	if err != nil {
		if v, ok := err.(*jwt.ValidationError); ok {
			if v.Errors&jwt.ValidationErrorExpired == 0 {
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
