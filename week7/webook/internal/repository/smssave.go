package repository

import (
	"context"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"time"
)

type SmsRepository interface {
	Save(ctx context.Context, sms domain.Sms) error
	Fetch(ctx context.Context) (domain.Sms, error)
}

type SaveSmsRepository struct {
	dao *dao.GORMSmsDao
}

func NewSaveSmsRepository(dao *dao.GORMSmsDao) SmsRepository {
	return &SaveSmsRepository{
		dao: dao,
	}
}

func (r *SaveSmsRepository) Save(ctx context.Context, sms domain.Sms) error {
	return r.dao.Add(ctx, dao.Sms{
		TplId: sms.TplId,
		Phone: sms.Phone,
		Msg:   sms.Msg,
	})
}

func (r *SaveSmsRepository) Fetch(ctx context.Context) (domain.Sms, error) {
	s, err := r.dao.GetAndDel(ctx)
	if err != nil {
		return domain.Sms{}, err
	}
	return entityToDomain(s), nil
}

func entityToDomain(s dao.Sms) domain.Sms {
	unixMillis := s.Ctime
	seconds := unixMillis / 1000
	nanos := (unixMillis % 1000) * 1e6

	t := time.Unix(seconds, nanos)
	return domain.Sms{
		TplId: s.TplId,
		Phone: s.Phone,
		Msg:   s.Msg,
		Id:    s.Id,
		Ctime: t,
	}
}
