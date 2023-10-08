package failover

import (
	"context"
	"errors"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"gitee.com/geekbang/basic-go/webook/internal/service/sms/ratelimit"
	"github.com/ecodeclub/ekit/retry"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strings"
	"sync/atomic"
	"time"
)

var (
	ErrSyncSave = errors.New("因发送失败转异步存储发送")
)

type SyncSaveFailoverSMSService struct {
	repo            repository.SmsRepository
	svc             ratelimit.RatelimitSMSService
	expiredtime     int   //信息失效时间，超过此时间信息失去效用，丢弃。(秒)
	interval        int   //异步调用发送的时间间隔。(秒)
	isworkerrunning int32 //0 表示未启动worker，1表示已启动worker
	needSave        int32 //0 表示不需要转存，1 表示需要转存
	retry           func() retry.Strategy
}

func NewSyncSaveFailoverSMSService(repo repository.SmsRepository,
	svc ratelimit.RatelimitSMSService, interval int,
	expiredtime int, retry func() retry.Strategy) *SyncSaveFailoverSMSService {
	return &SyncSaveFailoverSMSService{
		repo:            repo,
		svc:             svc,
		expiredtime:     expiredtime,
		interval:        interval,
		isworkerrunning: 0,
		needSave:        0,
		retry:           retry,
	}
}

func entityToDoMain(tplId string, args []string, numbers ...string) domain.Sms {
	var sms domain.Sms

	sms.TplId = tplId
	allnumbers := strings.Join(numbers, "|")
	sms.Phone = allnumbers
	sms.Msg = strings.Join(args, "|")
	return sms
}

func doMainToEntity(sms domain.Sms) (string, []string, []string) {
	tplId := sms.TplId
	allnumbers := strings.Split(sms.Phone, "|")
	allArgs := strings.Split(sms.Msg, "|")
	return tplId, allArgs, allnumbers
}

func (s *SyncSaveFailoverSMSService) writeNeedSave(isNeedSave int32) {
	atomic.StoreInt32(&s.needSave, isNeedSave)
}
func (s *SyncSaveFailoverSMSService) readNeedSave() int32 {
	return atomic.LoadInt32(&s.needSave)
}

// 异步处理短信消息
func (s *SyncSaveFailoverSMSService) worker(ctx context.Context) {
	succeedCnt := 0
	totalCnt := 0
	for {
		//	从数据库读取消息并发送
		sms, err := s.repo.Fetch(ctx)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				//所有数据库处理完毕
				if succeedCnt != 0 && succeedCnt == totalCnt {
					//表示服务商已正常，通知不用转存
					s.writeNeedSave(0)
				}
				succeedCnt = 0
				totalCnt = 0
				select {
				case <-ctx.Done():
					return
				}
				time.Sleep(time.Second * time.Duration(s.interval))
			} else {
				//数据库可能崩溃，告警通知dba介入检查
				zap.L().Warn("从数据库读取短信消息时数据库出错:", zap.Error(err))
				select {
				case <-ctx.Done():
					return
				}
				time.Sleep(time.Second * time.Duration(s.interval))
			}
			continue
		}
		totalCnt++
		now := time.Now()
		//过期信息不用发送
		if now.Sub(sms.Ctime).Seconds() > float64(s.expiredtime) {
			succeedCnt++
			continue
		}
		tplId, args, numbers := doMainToEntity(sms)
		err = s.svc.Send(ctx, tplId, args, numbers...)
		if err != nil {
			//还是失败，继续转存
			err := s.repo.Save(ctx, sms)
			if err != nil {
				zap.L().Warn("短信转存失败!", zap.Error(err))
			}
			continue
		}
		succeedCnt++
	}
}

func (s *SyncSaveFailoverSMSService) toSaveMsg(ctx context.Context, tplId string, args []string, numbers ...string) error {
	err := s.repo.Save(ctx, entityToDoMain(tplId, args, numbers...))
	if err != nil {
		zap.L().Warn("短信保存失败!", zap.Error(err))
		return err
	}
	return ErrSyncSave
}

func (s *SyncSaveFailoverSMSService) startWorker(ctx context.Context) {
	if atomic.CompareAndSwapInt32(&s.isworkerrunning, 0, 1) {
		go s.worker(ctx)
	}
}

func (s *SyncSaveFailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	needSave := s.readNeedSave()
	if needSave == 1 {
		return s.toSaveMsg(ctx, tplId, args, numbers...)
	}
	var retrytimer *time.Timer
	retry := s.retry()
	for {
		err := s.svc.Send(ctx, tplId, args, numbers...)
		if err == nil {
			return nil
		}
		//如果触发限流
		if err == ratelimit.ErrLimited {
			s.startWorker(ctx)
			return s.toSaveMsg(ctx, tplId, args, numbers...)
		} else if err != nil {
			timeInterval, try := retry.Next()
			//超过重试次数，表示对方服务可能崩溃，转存异步发送
			if !try {
				s.writeNeedSave(1)
				s.startWorker(ctx)
				return s.toSaveMsg(ctx, tplId, args, numbers...)
			}
			if retrytimer == nil {
				retrytimer = time.NewTimer(timeInterval)
			} else {
				retrytimer.Reset(timeInterval)
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-retrytimer.C:
			}
		}
	}
}
