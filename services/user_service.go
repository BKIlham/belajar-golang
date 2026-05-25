package services

import (
	"cobago/models"
	"cobago/repositories"
)

type UserService interface {
	RegisterUser(name, email, avatarURL string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	UpdateUser(id uint, name, email, avatarURL string) (*models.User, error)
	SoftDeleteUser(id uint) error
	HardDeleteUser(id uint) error // <-- Kita kembalikan ke format standar error saja
	RestoreUser(id uint) error
	GetAllUsers(page, limit int) ([]models.User, error)
}

type userServiceImpl struct {
	userRepo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userServiceImpl{userRepo: repo}
}

func (s *userServiceImpl) RegisterUser(name, email, avatarURL string) (*models.User, error) {
	user := &models.User{Name: name, Email: email, AvatarURL: avatarURL}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userServiceImpl) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *userServiceImpl) UpdateUser(id uint, name, email, avatarURL string) (*models.User, error) {
	// 1. Validasi apakah user-nya ada di DB
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 2. Isi ID beserta field yang mau di-update ke dalam satu objek struct
	// Properti string yang kosong ("") otomatis diabaikan oleh operasi .Updates() GORM di Repo
	updateFields := models.User{
		ID:        id, // Masukkan ID ke sini agar GORM tahu baris mana yang mau di-PATCH
		Name:      name,
		Email:     email,
		AvatarURL: avatarURL,
	}

	// 3. Lempar hanya 1 variabel struct pointer ke Repo sesuai blueprint interface terbaru
	if err := s.userRepo.Updates(&updateFields); err != nil {
		return nil, err
	}

	// 4. Ambil dan kembalikan data paling mutakhir dari database
	return s.userRepo.GetByID(id)
}

func (s *userServiceImpl) SoftDeleteUser(id uint) error {
	return s.userRepo.Delete(id)
}

func (s *userServiceImpl) HardDeleteUser(id uint) error {
	return s.userRepo.HardDelete(id)
}

func (s *userServiceImpl) RestoreUser(id uint) error {
	return s.userRepo.Restore(id)
}

func (s *userServiceImpl) GetAllUsers(page, limit int) ([]models.User, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	return s.userRepo.FindAll(page, limit)
}