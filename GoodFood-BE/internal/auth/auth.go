package auth

import (
	"GoodFood-BE/internal/dto"
	redisdatabase "GoodFood-BE/internal/redis-database"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Used to sign token and verify token
var secretKey = []byte(os.Getenv("JWT_SECRET"))

// Struct that stores info in JWT
// Username is to save into c.Locals().
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// HashToken returns a SHA256 hex representation of a token string.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// Used to create accessToken (expires in 15mins) and refreshToken (with jti) returns them plus the jti.
func CreateToken(username string) (accessTokenStr, refreshTokenStr, refreshJTI string, err error) {

	//Creating accessToken 15mins
	accessTokenClaims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString(secretKey)
	if err != nil {
		return "", "", "", err
	}

	//Refresh token with jti
	jti := uuid.New().String()
	refreshClaims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err = refreshToken.SignedString(secretKey)
	if err != nil {
		return "", "", "", err
	}

	return accessTokenString, refreshTokenStr, jti, nil
}

// VerifyToken parses and validates a JWT string and returns Claims.
func VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// SaveRefreshToken stores a new refresh token in Redis and adds jti to user's set.
// Key pattern: refresh:{jti} -> JSON(refreshRecord) with TTL = expiresAt - now.
// Also maintain a set user_refresh:{username} with members = jti (to support revoke-all or list sessions).
func SaveRefreshToken(jti, refreshToken, username, userAgent, ip string, expiresAt time.Time) error {
	hash := HashToken(refreshToken)
	rec := dto.RefreshRecord{
		Username:  username,
		TokenHash: hash,
		ExpiresAt: expiresAt,
		UserAgent: userAgent,
		IP:        ip,
	}
	b, _ := json.Marshal(rec)
	key := "refresh:" + jti
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		ttl = time.Second * 1
	}

	// Set the refresh record with TTL
	if err := redisdatabase.Client.Set(redisdatabase.Ctx, key, b, ttl).Err(); err != nil {
		return err
	}
	// Add jti to user's set
	setKey := "user_refresh:" + username
	if err := redisdatabase.Client.SAdd(redisdatabase.Ctx, setKey, jti).Err(); err != nil {
		// best effort: if set add fails, remove the refresh key
		_ = redisdatabase.Client.Del(redisdatabase.Ctx, key).Err()
		return err
	}
	// Set same TTL on the set membership? Redis sets don't have per-member TTL. We can set expire on set key to be >= token TTL.
	_ = redisdatabase.Client.Expire(redisdatabase.Ctx, setKey, time.Hour*24*8).Err() // keep set around; optional: tune
	return nil
}

// GetRefreshRecord loads refresh record by jti. Returns nil if not found.
func GetRefreshRecord(jti string) (*dto.RefreshRecord, error) {
	key := "refresh:" + jti
	val, err := redisdatabase.Client.Get(redisdatabase.Ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var rec dto.RefreshRecord
	if err := json.Unmarshal([]byte(val), &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// RevokeRefreshToken deletes the refresh token record and removes jti from user's set.
func RevokeRefreshToken(jti string) error {
	key := "refresh:" + jti
	// try to get username before deletion to remove from user set
	val, _ := redisdatabase.Client.Get(redisdatabase.Ctx, key).Result()
	if val != "" {
		var rec dto.RefreshRecord
		_ = json.Unmarshal([]byte(val), &rec)
		_ = redisdatabase.Client.SRem(redisdatabase.Ctx, "user_refresh:"+rec.Username, jti).Err()
	}
	return redisdatabase.Client.Del(redisdatabase.Ctx, key).Err()
}

// RevokeAllUserRefreshTokens revokes (deletes) all sessions for a username.
func RevokeAllUserRefreshTokens(username string) error {
	setKey := "user_refresh:" + username
	jtis, err := redisdatabase.Client.SMembers(redisdatabase.Ctx, setKey).Result()
	if err != nil {
		return err
	}
	pipe := redisdatabase.Client.TxPipeline()
	for _, j := range jtis {
		pipe.Del(redisdatabase.Ctx, "refresh:"+j)
		pipe.SRem(redisdatabase.Ctx, setKey, j)
	}
	_, err = pipe.Exec(redisdatabase.Ctx)
	return err
}

// ValidateRefreshAndRotate verifies incoming refresh token, checks Redis record, rotates tokens.
func ValidateRefreshAndRotate(refreshToken, userAgent, ip string) (newAccess string, newRefresh string, newJTI string, err error) {
	claims, err := VerifyToken(refreshToken)
	if err != nil {
		return "", "", "", err
	}
	jti := claims.ID
	if jti == "" {
		return "", "", "", errors.New("missing jti in refresh token")
	}
	rec, err := GetRefreshRecord(jti)
	if err != nil {
		return "", "", "", err
	}
	if rec == nil {
		return "", "", "", errors.New("refresh token not found or revoked")
	}
	if rec.TokenHash != HashToken(refreshToken) {
		_ = RevokeAllUserRefreshTokens(rec.Username)
		return "", "", "", errors.New("refresh token reuse detected")
	}

	newAccess, newRefresh, newJti, err := CreateToken(claims.Username)
	if err != nil {
		return "", "", "", err
	}
	newClaims, _ := VerifyToken(newRefresh)
	_ = SaveRefreshToken(newJti, newRefresh, claims.Username, userAgent, ip, newClaims.ExpiresAt.Time)
	_ = RevokeRefreshToken(jti)
	return newAccess, newRefresh, newJti, nil
}
