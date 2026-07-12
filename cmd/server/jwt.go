package main

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ログインしたユーザー情報をJWTトークンに格納するクレーム
type LoginClaim struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func generateJWT(userID int, role string, secret string) (string, error) {
	// 秘密鍵が空文字列の場合、エラーを返す
	if secret == "" {
		return "", errors.New("JWT_SECRET is not set")
	}

	now := time.Now()

	claims := LoginClaim{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),                     // 発行日時
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), // 有効期限（24時間）
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	// JWTトークンを秘密鍵で署名して文字列として返す
	return token.SignedString([]byte(secret))
}
