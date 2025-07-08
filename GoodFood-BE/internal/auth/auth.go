package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("secret-key")

//Struct chứa thông tin trong JWT
type Claims struct{
	Username string `json:"username"`
	jwt.RegisteredClaims
}

//Hàm tạo accessToken 15p và refreshToken 7 ngày
func CreateToken(username string) (string,string,error){
	//Tạo accessToken 15p
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
		return "","",err
	}

	//Tạo refreshToken 7 ngày
	refreshTokenClaims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256,refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString(secretKey)
	if err != nil{
		return "","",err
	}

	return accessTokenString, refreshTokenString, nil
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