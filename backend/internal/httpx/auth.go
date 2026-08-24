package httpx

import (
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func SignJWT(secret string, userID int64, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	c := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Subject:   strconv.FormatInt(userID, 10),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString([]byte(secret))
}

func ParseJWT(secret, token string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, Unauthorized("无效令牌")
		}
		return []byte(secret), nil
	})
	if err != nil || !t.Valid {
		return nil, Unauthorized("未登录或令牌已过期")
	}
	c, ok := t.Claims.(*Claims)
	if !ok || c.UserID == 0 {
		return nil, Unauthorized("无效令牌")
	}
	return c, nil
}

func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := Bearer(r)
			if raw == "" {
				Fail(w, r, Unauthorized("请先登录"))
				return
			}
			c, err := ParseJWT(secret, raw)
			if err != nil {
				Fail(w, r, err)
				return
			}
			ctx := WithUser(r.Context(), c.UserID, c.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Role(r) != "admin" {
			Fail(w, r, Forbidden("需要管理员权限"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
