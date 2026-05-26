package services

import (
	"cobago/repositories"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, string, error)
	Logout(ctx context.Context, userID uint) error
	Refresh(ctx context.Context, refreshToken string) (string, error)
}

type authServiceImpl struct {
	userRepo repositories.UserRepository
	redis    *redis.Client
}

func NewAuthService(repo repositories.UserRepository, rdb *redis.Client) AuthService {
	return &authServiceImpl{userRepo: repo, redis: rdb}
}

func (s *authServiceImpl) Login(ctx context.Context, email, password string) (string, string, error) {
	// 1. Cari user di DB berdasarkan email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", "", errors.New("email atau password salah")
	}

	// 2. Bandingkan password mentah dengan hash bcrypt di DB
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", errors.New("email atau password salah")
	}

	// 3. Generate ACCESS TOKEN (Umur pendek: 15 Menit)
	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := accessTokenObj.SignedString([]byte(os.Getenv("JWT_ACCESS_SECRET")))
	if err != nil {
		return "", "", err
	}

	// 4. Generate REFRESH TOKEN (Umur panjang: 7 Hari)
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenObj.SignedString([]byte(os.Getenv("JWT_REFRESH_SECRET")))
	if err != nil {
		return "", "", err
	}

	// 5. Simpan Refresh Token ke Redis (Whitelist Session aktif)
	redisKey := fmt.Sprintf("refresh_token:%d", user.ID)
	err = s.redis.Set(ctx, redisKey, refreshToken, 7*24*time.Hour).Err()
	if err != nil {
		return "", "", errors.New("gagal membuat sesi login di server")
	}

	return accessToken, refreshToken, nil
}

func (s *authServiceImpl) Logout(ctx context.Context, userID uint) error {
	// Hapus token dari Redis agar tidak bisa di-refresh lagi
	redisKey := fmt.Sprintf("refresh_token:%d", userID)
	return s.redis.Del(ctx, redisKey).Err()
}

func (s *authServiceImpl) Refresh(ctx context.Context, refreshTokenStr string) (string, error) {
	// 1. Parse dan validasi Refresh Token string
	token, err := jwt.Parse(refreshTokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_REFRESH_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("refresh token tidak valid atau kedaluwarsa")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("gagal membaca data token")
	}

	userID := uint(claims["user_id"].(float64))

	// 2. Validasi apakah Refresh Token tersebut masih terdaftar di Redis
	redisKey := fmt.Sprintf("refresh_token:%d", userID)
	savedToken, err := s.redis.Get(ctx, redisKey).Result()
	if err != nil || savedToken != refreshTokenStr {
		return "", errors.New("sesi login telah berakhir, silakan login ulang")
	}

	// 3. Jika valid, buatkan ACCESS TOKEN baru yang segar
	newAccessClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}
	newAccessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, newAccessClaims)
	newAccessToken, err := newAccessTokenObj.SignedString([]byte(os.Getenv("JWT_ACCESS_SECRET")))
	if err != nil {
		return "", nil
	}

	return newAccessToken, nil
}