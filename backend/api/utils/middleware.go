package utils

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	Username string `json:"username"`
	UserID   int64  `json:"userid"`
	jwt.RegisteredClaims
}

func JWTMiddleWare() func(gin *gin.Context) {
	return func(c *gin.Context) {
		jtoken, err := extractToken(c)
		if err != nil || jtoken == "" {
			Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}
		var temp JWT
		_, err = jwt.ParseWithClaims(jtoken, &temp, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("secret"), nil
		})
		if err != nil {
			Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}
		c.Set("userid", temp.UserID)
		c.Set("username", temp.Username)
		c.Next()
	}
}

func extractToken(ctx *gin.Context) (string, error) {
	auth := ctx.GetHeader("authorization")
	if auth != "" {
		strArr := strings.Split(auth, " ")
		if len(strArr) == 2 && strArr[0] == "Bearer" {
			return strArr[1], nil
		}
		return "", errors.New("invalid token")
	}
	return "", errors.New("token is not found")
}

func GenerateToken(userid int64, username string) (string, error) {
	temp := JWT{
		Username: username,
		UserID:   userid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, temp)
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func CorsMiddleWare() func(gin *gin.Context) {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
