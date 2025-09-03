package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

//Used to sign token and verify token
var secretKey = []byte(os.Getenv("JWT_SECRET"))

//Struct that stores info in JWT
//Username is to save into c.Locals(). SessionID is to differentiate the logged in users 
type Claims struct{
	Username string `json:"username"`
	SessionID string `json:"sessionID"`
	jwt.RegisteredClaims
}

//SessionID generation is for differentiating the logged in users
func generateSessionID() string{
	bytes := make([]byte,16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

//Used to create accessToken (expires in 15mins) and refreshToken (expires in 7 days)
func CreateToken(username string, sessionID string) (string,string,string,error){
	//Creating sessionID for every login session
	if sessionID == ""{
		sessionID = generateSessionID();
	}

	//Creating accessToken 15mins
	accessTokenClaims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
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

//Verify token and return struct Claims if token is valid
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