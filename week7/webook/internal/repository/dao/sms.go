package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type Sms struct {
	Id    int64 `gorm:"primaryKey,autoIncrement"`
	TplId string
	Phone string
	Msg   string
	// 创建时间
	Ctime int64
}

type SmsDao interface {
	Add(ctx context.Context, sms Sms) error
	GetAndDel(ctx context.Context) (Sms, error)
}

type GORMSmsDao struct {
	db *gorm.DB
}

func NewGORMSmsDao(db *gorm.DB) *GORMSmsDao {
	return &GORMSmsDao{db: db}
}

func (s *GORMSmsDao) Add(ctx context.Context, sms Sms) error {
	if sms.Ctime == 0 {
		sms.Ctime = time.Now().UnixMilli()
	}
	return s.db.WithContext(ctx).Create(sms).Error
}

func (s *GORMSmsDao) GetAndDel(ctx context.Context) (Sms, error) {
	var sms Sms
	err := s.db.WithContext(ctx).First(sms).Error
	if err != nil {
		return Sms{}, err
	}
	id := sms.Id
	err = s.db.WithContext(ctx).Where("id = ?", id).Delete(sms).Error
	return sms, err
}
