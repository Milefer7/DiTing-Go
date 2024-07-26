package utils

import (
	"DiTing-Go/global"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"time"
)

// 从配置文件中获取JWT的密钥
var jwtSecret = []byte(viper.GetString("jwt.secret"))

// Claims结构体用于JWT的载荷部分
type Claims struct {
	Uid                int64 `json:"uid"` // 用户ID
	jwt.StandardClaims       // 标准声明
}

// GenerateToken 生成token
// uid 用户ID
// 返回生成的token字符串和错误信息
func GenerateToken(uid int64) (string, error) {
	nowTime := time.Now()                           // 当前时间
	expireTime := nowTime.Add(24 * 360 * time.Hour) // 过期时间为360天后

	claims := Claims{
		uid,
		jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(), // 过期时间
			Issuer:    "diting-go",       // 签发者
		},
	}

	// 使用指定的签名方法和载荷生成token
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret) // 使用密钥签名token
	if err != nil {
		global.Logger.Errorf("generate token failed: %v", err) // 记录生成token失败的错误
	}
	return token, err
}

// ParseToken 解析token
// tokenString 待解析的token字符串
// 返回解析后的Claims和错误信息
func ParseToken(tokenString string) (*Claims, error) {
	// 解析token并验证签名
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (i interface{}, err error) {
		return jwtSecret, nil // 返回密钥进行签名验证
	})
	if err != nil {
		return nil, err
	}
	// 校验token是否有效
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token") // 返回无效token的错误
}
