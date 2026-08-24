package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"gocamping/internal/httpx"
	"gocamping/internal/model"
	"gocamping/internal/repo"
	"gocamping/internal/timeutil"
)

func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

func Register(ctx context.Context, users *repo.UserRepo, username, password, nickname string) (*model.User, error) {
	username = strings.TrimSpace(username)
	nickname = strings.TrimSpace(nickname)
	if utf8.RuneCountInString(username) < 3 || utf8.RuneCountInString(username) > 32 {
		return nil, httpx.Validation("用户名长度 3–32")
	}
	if len(password) < 6 {
		return nil, httpx.Validation("密码至少 6 位")
	}
	if nickname == "" {
		nickname = username
	}
	exist, err := users.ByUsername(ctx, username)
	if err != nil {
		return nil, httpx.Internal("查询用户失败")
	}
	if exist != nil {
		return nil, httpx.Conflict("用户名已存在")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return nil, httpx.Internal("密码加密失败")
	}
	u := &model.User{Username: username, PasswordHash: hash, Nickname: nickname, Role: "member", CreatedAt: timeutil.NowNaive()}
	if err := users.Create(ctx, u); err != nil {
		return nil, httpx.Internal("创建用户失败")
	}
	return u, nil
}
