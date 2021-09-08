package email

import (
	"github.com/volam1999/gomail/internal/app/types"
	"gorm.io/gorm"
)

type MysqlDBRepository struct {
	db *gorm.DB
}

func NewMysqlDBRepository(db *gorm.DB) *MysqlDBRepository {
	return &MysqlDBRepository{
		db: db,
	}
}

func (r *MysqlDBRepository) Create(email *types.Email) (int, error) {
	result := r.db.Create(email)
	if result.Error != nil {
		return -1, result.Error
	}
	return email.Id, nil
}

func (r *MysqlDBRepository) Update(emailId int, email *types.Email) error {
	err := r.db.Model(&types.Email{}).Where("Id = ?", emailId).Updates(email).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *MysqlDBRepository) FindAllScheduleEmail() (*[]types.Email, error) {
	var emails []types.Email
	err := r.db.Where("status = ?", "PENDING").Find(&emails).Error
	if err != nil {
		return nil, err
	}
	return &emails, nil
}

func (r *MysqlDBRepository) FindAll() (*[]types.Email, error) {
	var emails []types.Email
	result := r.db.Find(&emails)
	if result.Error != nil {
		return nil, result.Error
	}
	return &emails, nil
}

func (r *MysqlDBRepository) FindByEmailId(emailId int) (*types.Email, error) {
	var email types.Email
	err := r.db.First(&email, emailId).Error
	if err != nil {
		return nil, err
	}
	return &email, nil
}
