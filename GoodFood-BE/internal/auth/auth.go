package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("secret-key")

//Struct chứa thông tin trong JWT
type Claims struct{
	Username string `json:"username"`
	SessionID string `json:"sessionID"`
	jwt.RegisteredClaims
}

func generateSessionID() string{
	bytes := make([]byte,16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

//used to create accessToken (expires in 15mins) and refreshToken (expires in 7 days)
func CreateToken(username string, sessionID string) (string,string,string,error){
	//creating sessionID for every login session
	if sessionID == ""{
		sessionID = generateSessionID();
	}

	//Creating accessToken 15mins
	accessTokenClaims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Second)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256,accessTokenClaims)
	accessTokenString, err := accessToken.SignedString(secretKey)
	if err != nil{
		return "","","",err
	}

	//refreshToken 7 days
	refreshTokenClaims := Claims{
		Username: username,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256,refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString(secretKey)
	if err != nil{
		return "","","",err
	}

	return accessTokenString, refreshTokenString,sessionID, nil
}

//Xác thực token và trả về Claims nếu hợp lệ
func VerifyToken(tokenString string) (*Claims, error){
	token, err := jwt.ParseWithClaims(tokenString,&Claims{}, func(token *jwt.Token) (interface{}, error){
		return secretKey, nil
	})
	if err != nil{
		return nil,err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid{
		return nil, errors.New("invalid token")
	}

	return claims,nil
}