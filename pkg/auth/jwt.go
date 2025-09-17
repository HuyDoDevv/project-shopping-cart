package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"gin/user-management-api/internal/db/sqlc"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/pkg/cache"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	cache *cache.RedisCacheService
}

// type Claim struct {
// 	jwt.RegisteredClaims
// }

type EncryptedPayload struct {
	UserUUID string `json:"user_uuid"`
	Email string `json:"email"`
	Role int32 `json:"role"`
}

type RefreshToken struct {
	Token string `json:"token"`
	UserUUID string `json:"user_uuid"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked bool `json:"revoked"`
}

var (
	jwtSecret = []byte(utils.GetEnv("JWT_SECRET", "JWT-SECRET-HUY-DO-DANG-HOC-GOLANG-CUC-CHILL"))
	jwtEncrypKey = []byte(utils.GetEnv("JWT_ENCRYPT_KEY", "12345678901234567890123456789027"))
)

const (
	AccessTokenTTL = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

func NewJWTService(cache *cache.RedisCacheService) TokenService{
	return &JWTService{
		cache: cache,
	}
}


func (js *JWTService) GenerateAccessToken(user sqlc.User) (string, error) {
	payload := &EncryptedPayload{
		UserUUID: user.UserUuid.String(),
		Email: user.UserEmail,
		Role: user.UserLevel,
	}

	rawData, err := json.Marshal(payload)
	if err != nil {
		return "", nil
	}

	encrypted, err := utils.EncryptAES(rawData, jwtEncrypKey)
	if err != nil {
		return "", nil
	}

	claims := jwt.MapClaims{
		"data": encrypted,
		"jti": uuid.NewString(),
		"exp": jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
		"iat": jwt.NewNumericDate(time.Now()),
		"iss": "HuyDo",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func (js *JWTService) ParseToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error){
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, nil, utils.NewError(utils.UnauthorizedError, "Invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, utils.NewError(utils.UnauthorizedError, "Invalid claims")

	}

	return token, claims, nil
}

func (js *JWTService) DecryptAccessTokenPayload(tokenString string) (*EncryptedPayload, error) {
	_, claims, err := js.ParseToken(tokenString)
	if err != nil {
		return nil, utils.WrapError(utils.UnauthorizedError, "Cannot parse token", err)
	}

	enctyptedData, ok := claims["data"].(string)
	if !ok {
		return nil, utils.WrapError(utils.InternalServerError, "Encoded data not found", err)
	}

	decryptedBytes, err := utils.DecryptAES(enctyptedData, jwtEncrypKey)
	if err != nil {
		return nil, utils.WrapError(utils.UnauthorizedError, "Cannot parse token", err)
	}

	var payload EncryptedPayload
	if err := json.Unmarshal(decryptedBytes, &payload); err != nil {
		return nil, utils.WrapError(utils.UnauthorizedError, "Invalid data format", err)
	}
	return &payload, nil
}


func (js *JWTService) GenerateRefreshToken(user sqlc.User) (RefreshToken, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return RefreshToken{}, err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)


	return RefreshToken{
		Token: token,
		UserUUID: user.UserUuid.String(),
		ExpiresAt: time.Now().Add(RefreshTokenTTL),
		Revoked: false,
	}, nil
}


func (js *JWTService) StoreRefreshToken(token RefreshToken) error {
	cacheKey := "refresh_token:" + token.Token
	return js.cache.Set(cacheKey, token, RefreshTokenTTL)
}


func (js *JWTService) ValidateRefreshToken(token string) (RefreshToken, error) {
	cacheKey := "refresh_token:" + token
	var refreshToken RefreshToken

	err := js.cache.Get(cacheKey, &refreshToken)
	if err != nil || refreshToken.Revoked || refreshToken.ExpiresAt.Before(time.Now()) {
		return RefreshToken{}, utils.WrapError(utils.InternalServerError, "Cannot get refresh token", err)
	}
	return refreshToken, nil
}

func (js *JWTService) RevokeRefreshToken(token string) error {
	cacheKey := "refresh_token:" + token
	var refreshToken RefreshToken

	err := js.cache.Get(cacheKey, &refreshToken)
	if err != nil {
		return utils.WrapError(utils.InternalServerError, "Cannot get refresh token", err)
	}
	refreshToken.Revoked = true
	return js.cache.Set(cacheKey, refreshToken, time.Until(refreshToken.ExpiresAt))
}
