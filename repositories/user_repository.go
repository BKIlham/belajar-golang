package repositories

import (
	"cobago/models"
	"errors"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	Updates(user *models.User) error
	Delete(id uint) error
	HardDelete(id uint) error
	Restore(id uint) error
	FindAll(page, limit int) ([]models.User, error)
	GetByEmail(email string) (*models.User, error)
}

type userRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepositoryImpl) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) Updates(user *models.User) error {
	return r.db.Model(user).Updates(user).Error
}

func (r *userRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepositoryImpl) HardDelete(id uint) error {
	return r.db.Unscoped().Delete(&models.User{}, id).Error
}

func (r *userRepositoryImpl) Restore(id uint) error {
	var user models.User
	if err := r.db.Unscoped().First(&user, id).Error; err != nil {
		return err
	}
	return r.db.Unscoped().Model(&user).UpdateColumn("deleted_at", nil).Error
}

func (r *userRepositoryImpl) FindAll(page, limit int) ([]models.User, error) {
	var users []models.User

	// Hitung offset (data ke berapa yang mulai diambil)
	// Misal page 2, limit 10 -> offset = (2-1) * 10 = 10 (lewati 10 data pertama)
	offset := (page - 1) * limit

	// Tarik data dengan batasan limit dan offset
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepositoryImpl) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("email atau password salah")
		}
		return nil, err
	}
	return &user, nil
}