package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("video_feed_secret_key") // 生产环境应改为配置

// Claims 自定义JWT载荷
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT Token，有效期24小时
func GenerateToken(userID uint, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "video_feed",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析并验证JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("无效的Token")
	}

	return claims, nil
}

// JWTAuth 鉴权中间件(上面的都只是辅助函数，这里才是中间件的核心逻辑)，检查用户是否携带了有效的JWT通行证
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 第一步：检查有没有带"会员卡"
		// 从请求头获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "缺少Token",
			})
			c.Abort() // 阻止调用后续的路由处理函数
			return
		}

		// 第二步：检查"会员卡"格式对不对（必须是 Bearer + 空格 + JWT）
		// 格式: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Token格式错误",
			})
			c.Abort()
			return
		}

		// 第三步：验证"会员卡"真伪（解析JWT）
		// 解析Token
		claims, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Token无效或已过期",
			})
			c.Abort()
			return
		}

		// 第四步：验证通过，把用户信息记下来，方便后续使用
		// 将用户信息存入上下文，后续处理函数可以直接使用
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		// 第五步：放行，继续处理真正的业务逻辑
		c.Next()
	}
}
