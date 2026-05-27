package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

// User 用户模型
type User struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	Username     string `gorm:"type:varchar(255);uniqueIndex;not null"`
	Email        string `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (d *UserDAO) Create(ctx context.Context, user *User) error {
	return d.db.WithContext(ctx).Create(user).Error
}

func (d *UserDAO) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := d.db.WithContext(ctx).Model(&User{}).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("email or password is incorrect")
		}
		return nil, err
	}
	return &user, nil
}
