package failover

import (
	"context"
	"errors"

	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	repomocks "gitee.com/geekbang/basic-go/webook/internal/repository/mocks"
	smsmocks "gitee.com/geekbang/basic-go/webook/internal/service/sms/mocks"
	"gitee.com/geekbang/basic-go/webook/internal/service/sms/ratelimit"
	"gitee.com/geekbang/basic-go/webook/pkg/ratelimit/mocks"
	"github.com/ecodeclub/ekit/retry"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"reflect"
	"testing"
	"time"
)

func Test_doMainToEntity(t *testing.T) {
	type args struct {
		sms domain.Sms
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []string
		want2 []string
	}{
		{
			name: "doMainToEntity",
			args: args{
				sms: domain.Sms{
					TplId: "123456",
					Phone: "18888888888|19999999999",
					Msg:   "这是一条测试短信|测试是否收到",
				},
			},
			want:  "123456",
			want1: []string{"这是一条测试短信", "测试是否收到"},
			want2: []string{"18888888888", "19999999999"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := doMainToEntity(tt.args.sms)
			if got != tt.want {
				t.Errorf("doMainToEntity() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("doMainToEntity() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("doMainToEntity() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_entityToDoMain(t *testing.T) {
	type args struct {
		tplId   string
		args    []string
		numbers []string
	}
	tests := []struct {
		name string
		args args
		want domain.Sms
	}{
		{
			name: "entityToDoMain",
			args: args{
				tplId:   "123456",
				args:    []string{"这是一条测试短信", "测试是否收到"},
				numbers: []string{"18888888888", "19999999999"},
			},
			want: domain.Sms{
				TplId: "123456",
				Phone: "18888888888|19999999999",
				Msg:   "这是一条测试短信|测试是否收到",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := entityToDoMain(tt.args.tplId, tt.args.args, tt.args.numbers...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("entityToDoMain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getRetry() retry.Strategy {
	f, _ := retry.NewFixedIntervalRetryStrategy(time.Second*2, 3)
	return f
}

// 消息通道的读写都测了
func TestSyncSaveFailoverSMSService_readNeedSave(t *testing.T) {
	testCases := []struct {
		name         string
		mock         func(*gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService)
		needSaveFlag int32
		want         int32
	}{
		{
			name: "readNeedSave-false",
			mock: func(ctrl *gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService) {
				repo := repomocks.NewMockSmsRepository(ctrl)
				smssvc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				ratelimiter := ratelimit.NewRatelimitSMSService(smssvc, limiter)
				return repo, *ratelimiter
			},
			needSaveFlag: 0,
			want:         0,
		},
		{
			name: "readNeedSave-true",
			mock: func(ctrl *gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService) {
				repo := repomocks.NewMockSmsRepository(ctrl)
				smssvc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				ratelimiter := ratelimit.NewRatelimitSMSService(smssvc, limiter)
				return repo, *ratelimiter
			},
			needSaveFlag: 1,
			want:         1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo, ratelimit := tc.mock(ctrl)
			svc := NewSyncSaveFailoverSMSService(repo, ratelimit, 60, 600, getRetry)
			svc.readNeedSave() //先清空channel内容
			svc.writeNeedSave(tc.needSaveFlag)
			svc.writeNeedSave(tc.needSaveFlag)
			svc.writeNeedSave(tc.needSaveFlag)
			got := svc.readNeedSave()
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestSyncSaveFailoverSMSService_Send(t *testing.T) {
	testCases := []struct {
		name         string
		ctx          context.Context
		mock         func(*gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService)
		needSaveFlag int32
		TplId        string
		Args         []string
		Phone        []string
		wantErr      error
	}{
		{
			name: "发送成功",
			ctx:  context.Background(),
			mock: func(ctrl *gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService) {
				repo := repomocks.NewMockSmsRepository(ctrl)
				smssvc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				smssvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				ratelimiter := ratelimit.NewRatelimitSMSService(smssvc, limiter)

				return repo, *ratelimiter
			},

			TplId: "123456",
			Phone: []string{"19999999999"},
			Args:  []string{"这是测试信息"},

			needSaveFlag: 0,
			wantErr:      nil,
		},
		{
			name: "直接转存异步发送",
			ctx:  context.Background(),
			mock: func(ctrl *gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService) {
				repo := repomocks.NewMockSmsRepository(ctrl)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
				smssvc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				ratelimiter := ratelimit.NewRatelimitSMSService(smssvc, limiter)
				return repo, *ratelimiter
			},

			TplId: "123456",
			Phone: []string{"19999999999"},
			Args:  []string{"这是测试信息"},

			needSaveFlag: 1,
			wantErr:      ErrSyncSave,
		},
		{
			name: "限流转存",
			ctx:  context.Background(),
			mock: func(ctrl *gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService) {
				repo := repomocks.NewMockSmsRepository(ctrl)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
				repo.EXPECT().Fetch(gomock.Any()).Return(domain.Sms{}, gorm.ErrRecordNotFound).MinTimes(0)
				smssvc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				ratelimiter := ratelimit.NewRatelimitSMSService(smssvc, limiter)
				return repo, *ratelimiter
			},

			TplId: "123456",
			Phone: []string{"19999999999"},
			Args:  []string{"这是测试信息"},

			needSaveFlag: 0,
			wantErr:      ErrSyncSave,
		},
		{
			name: "发送失败次数超标转存",
			ctx:  context.Background(),
			mock: func(ctrl *gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService) {
				repo := repomocks.NewMockSmsRepository(ctrl)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
				repo.EXPECT().Fetch(gomock.Any()).Return(domain.Sms{}, gorm.ErrRecordNotFound).MinTimes(0)
				smssvc := smsmocks.NewMockService(ctrl)
				smssvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("发送失败")).Times(4)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil).Times(4)
				ratelimiter := ratelimit.NewRatelimitSMSService(smssvc, limiter)
				return repo, *ratelimiter
			},

			TplId: "123456",
			Phone: []string{"19999999999"},
			Args:  []string{"这是测试信息"},

			needSaveFlag: 0,
			wantErr:      ErrSyncSave,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo, ratelimit := tc.mock(ctrl)
			svc := NewSyncSaveFailoverSMSService(repo, ratelimit, 60, 600, getRetry)
			svc.writeNeedSave(tc.needSaveFlag)
			err := svc.Send(tc.ctx, tc.TplId, tc.Args, tc.Phone...)
			assert.Equal(t, err, tc.wantErr)
		})
	}
}

func TestSyncSaveFailoverSMSService_worker(t *testing.T) {
	testCases := []struct {
		name     string
		ctx      context.Context
		mock     func(*gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService)
		wantErr  error
		wantflag int32
	}{
		{
			name: "正常发送后进入等待",
			ctx:  context.Background(),
			mock: func(ctrl *gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService) {
				repo := repomocks.NewMockSmsRepository(ctrl)
				repo.EXPECT().Fetch(gomock.Any()).Return(domain.Sms{
					Id:    1,
					TplId: "123456",
					Phone: "19999999999",
					Msg:   "测试短信",
					Ctime: time.Now(),
				}, nil).Times(1)
				repo.EXPECT().Fetch(gomock.Any()).Return(domain.Sms{
					Id:    1,
					TplId: "123456",
					Phone: "19999999999",
					Msg:   "测试短信",
					Ctime: time.Now().Add(-time.Hour),
				}, nil).Times(1)
				repo.EXPECT().Fetch(gomock.Any()).Return(domain.Sms{}, gorm.ErrRecordNotFound).Times(1)
				smssvc := smsmocks.NewMockService(ctrl)
				smssvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).MaxTimes(2)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil).MaxTimes(2)
				ratelimiter := ratelimit.NewRatelimitSMSService(smssvc, limiter)
				return repo, *ratelimiter
			},
			wantflag: 0,
			wantErr:  nil,
		},
		{
			name: "发送失败继续转存",
			ctx:  context.Background(),
			mock: func(ctrl *gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService) {
				repo := repomocks.NewMockSmsRepository(ctrl)
				repo.EXPECT().Fetch(gomock.Any()).Return(domain.Sms{
					Id:    1,
					TplId: "123456",
					Phone: "19999999999",
					Msg:   "测试短信",
					Ctime: time.Now(),
				}, nil).Times(1)
				repo.EXPECT().Fetch(gomock.Any()).Return(domain.Sms{}, gorm.ErrRecordNotFound).Times(1)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				smssvc := smsmocks.NewMockService(ctrl)
				smssvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("发送失败"))
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				ratelimiter := ratelimit.NewRatelimitSMSService(smssvc, limiter)
				return repo, *ratelimiter
			},
			wantflag: 1,
			wantErr:  nil,
		},
		{
			name: "发送失败继续转存也失败",
			ctx:  context.Background(),
			mock: func(ctrl *gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService) {
				repo := repomocks.NewMockSmsRepository(ctrl)
				repo.EXPECT().Fetch(gomock.Any()).Return(domain.Sms{
					Id:    1,
					TplId: "123456",
					Phone: "19999999999",
					Msg:   "测试短信",
					Ctime: time.Now(),
				}, nil).Times(1)
				repo.EXPECT().Fetch(gomock.Any()).Return(domain.Sms{}, gorm.ErrRecordNotFound).Times(1)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("save 失败")).Times(1)
				smssvc := smsmocks.NewMockService(ctrl)
				smssvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("发送失败"))
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				ratelimiter := ratelimit.NewRatelimitSMSService(smssvc, limiter)
				return repo, *ratelimiter
			},
			wantflag: 1,
			wantErr:  nil,
		},
		{
			name: "数据库崩溃无法读取信息",
			ctx:  context.Background(),
			mock: func(ctrl *gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService) {
				repo := repomocks.NewMockSmsRepository(ctrl)
				repo.EXPECT().Fetch(gomock.Any()).Return(domain.Sms{}, errors.New("数据库崩溃啦")).Times(1)
				smssvc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				ratelimiter := ratelimit.NewRatelimitSMSService(smssvc, limiter)
				return repo, *ratelimiter
			},
			wantflag: 1,
			wantErr:  nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo, ratelimit := tc.mock(ctrl)
			svc := NewSyncSaveFailoverSMSService(repo, ratelimit, 60, 600, getRetry)
			svc.startWorker(tc.ctx)
			svc.writeNeedSave(1)
			time.Sleep(time.Second * 1)
			tc.ctx.Done()
			got := svc.readNeedSave()
			assert.Equal(t, tc.wantflag, got)
		})
	}
}

func TestSyncSaveFailoverSMSService_toSaveMsg(t *testing.T) {
	testCases := []struct {
		name    string
		ctx     context.Context
		tplId   string
		args    []string
		numbers []string
		mock    func(*gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService)
		wantErr error
	}{
		{
			name:    "短信转存数据库成功",
			ctx:     context.Background(),
			tplId:   "12345",
			args:    []string{"1", "2", "3", "4"},
			numbers: []string{"18888888888", "18888888888"},
			mock: func(ctrl *gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService) {
				repo := repomocks.NewMockSmsRepository(ctrl)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
				smssvc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				ratelimiter := ratelimit.NewRatelimitSMSService(smssvc, limiter)
				return repo, *ratelimiter
			},
			wantErr: ErrSyncSave,
		},
		{
			name:    "短信转存数据库失败",
			ctx:     context.Background(),
			tplId:   "12345",
			args:    []string{"1", "2", "3", "4"},
			numbers: []string{"18888888888", "18888888888"},
			mock: func(ctrl *gomock.Controller) (repository.SmsRepository, ratelimit.RatelimitSMSService) {
				repo := repomocks.NewMockSmsRepository(ctrl)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("转存失败"))
				smssvc := smsmocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				ratelimiter := ratelimit.NewRatelimitSMSService(smssvc, limiter)
				return repo, *ratelimiter
			},
			wantErr: errors.New("转存失败"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo, ratelimit := tc.mock(ctrl)
			svc := NewSyncSaveFailoverSMSService(repo, ratelimit, 60, 600, getRetry)
			goterr := svc.toSaveMsg(tc.ctx, tc.tplId, tc.args, tc.numbers...)
			assert.Equal(t, tc.wantErr, goterr)
		})
	}
}
