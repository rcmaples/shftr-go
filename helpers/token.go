package helpers

import (
	"encoding/json"
	"shftr/models"

	"github.com/golang-jwt/jwt/v4"
)


func UnmarshalToken(jwt jwt.MapClaims) models.Token {
	var token models.Token
	stringify, _ := json.Marshal(&jwt)
	json.Unmarshal([]byte(stringify), &token)
	return token
}