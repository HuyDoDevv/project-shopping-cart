package auth

import (
	"encoding/json"
	"gin/user-management-api/internal/db/sqlc"
	"gin/user-management-api/internal/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {

}

// type Claim struct {
// 	jwt.RegisteredClaims
// }

type EncryptedPayload struct {
	UserUUID string `json:"user_uuid"`
	Email string `json:"email"`
	Role int32 `json:"role"`
}

var (
	jwtSecret = []byte(utils.GetEnv("JWT_SECRET", "JWT-SECRET-HUY-DO-DANG-HOC-GOLANG-CUC-CHILL"))
	jwtEncrypKey = []byte(utils.GetEnv("JWT_ENCRYPT_KEY", "12345678901234567890123456789027"))
)

const (
	AccessTokenTTL = 15 * time.Minute
)

func NewJWTService() TokenService{
	return &JWTService{}
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

func (js *JWTService) GenerateRefreshToken() {

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
