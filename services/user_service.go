package services

import (
	"cobago/models"
	"cobago/repositories"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserService interface {
	RegisterUser(name, email, avatarURL string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	UpdateUser(id uint, name, email, avatarURL string) (*models.User, error)
	SoftDeleteUser(id uint) error
	HardDeleteUser(id uint) error
	RestoreUser(id uint) error
	GetAllUsers(ctx context.Context, page, limit int) ([]models.User, error)
}

type userServiceImpl struct {
	userRepo repositories.UserRepository
	redis    *redis.Client
}

func NewUserService(repo repositories.UserRepository, rdb *redis.Client) UserService {
	return &userServiceImpl{userRepo: repo, redis: rdb}
}

func (s *userServiceImpl) RegisterUser(name, email, avatarURL string) (*models.User, error) {
	user := &models.User{Name: name, Email: email, AvatarURL: avatarURL}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	s.clearUserCache()
	return user, nil
}

func (s *userServiceImpl) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *userServiceImpl) UpdateUser(id uint, name, email, avatarURL string) (*models.User, error) {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	updateFields := models.User{
		ID:        id,
		Name:      name,
		Email:     email,
		AvatarURL: avatarURL,
	}

	if err := s.userRepo.Updates(&updateFields); err != nil {
		return nil, err
	}

	s.clearUserCache()
	return s.userRepo.GetByID(id)
}

func (s *userServiceImpl) SoftDeleteUser(id uint) error {
	s.clearUserCache()
	return s.userRepo.Delete(id)
}

func (s *userServiceImpl) HardDeleteUser(id uint) error {
	s.clearUserCache()
	return s.userRepo.HardDelete(id)
}

func (s *userServiceImpl) RestoreUser(id uint) error {
	s.clearUserCache()
	return s.userRepo.Restore(id)
}

func (s *userServiceImpl) GetAllUsers(ctx context.Context, page, limit int) ([]models.User, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	cacheKey := fmt.Sprintf("users_cache_page_%d_limit_%d", page, limit)

	cachedData, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var users []models.User
		if json.Unmarshal([]byte(cachedData), &users) == nil {
			return users, nil
		}
	}

	users, err := s.userRepo.FindAll(page, limit)
	if err != nil {
		return nil, err
	}

	jsonData, errMarshal := json.Marshal(users)
	if errMarshal == nil {
		_ = s.redis.Set(ctx, cacheKey, jsonData, 5*time.Minute).Err()
	}

	return users, nil
}

func (s *userServiceImpl) clearUserCache() {
	ctx := context.Background()
	
	keys, err := s.redis.Keys(ctx, "users_cache_*").Result()
	if err != nil || len(keys) == 0 {
		return
	}

	_ = s.redis.Del(ctx, keys...).Err()
}