package repository

import (
	"example/ToDo/models"

	"gorm.io/gorm"
)

// 1. Interface
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByUsername(username string) (*models.User, error)
	UpgradeUserRole(userID uint, role string) error
}

// 2. Struct
type UserRepositoryImpl struct {
	DB *gorm.DB
}

// 3. Constructor
func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{DB: db}
}

// 4. Methods
func (r *UserRepositoryImpl) CreateUser(user *models.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepositoryImpl) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) UpgradeUserRole(userID uint, role string) error {
	return r.DB.Model(&models.User{}).Where("id = ?", userID).Update("role", role).Error
}
