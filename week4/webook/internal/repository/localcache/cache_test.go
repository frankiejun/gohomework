package localcache

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLocalCodeCache_Set(t *testing.T) {
	type args struct {
		ctx   context.Context
		biz   string
		phone string
		code  string
	}
	tests := []struct {
		name    string
		fields  *LocalCodeCache
		args    *args
		wantErr bool
	}{
		{
			name:   "settest",
			fields: newLocalCodeCache(60, 3, 600, 1).(*LocalCodeCache),
			args: &args{
				biz:   "codeCache",
				phone: "12345678911",
				code:  "12345",
			},
			wantErr: true,
		},

		{
			name:   "settest",
			fields: newLocalCodeCache(60, 3, 600, 1).(*LocalCodeCache),
			args: &args{
				biz:   "codeCache",
				phone: "12345678911",
				code:  "12345",
			},
			wantErr: true,
		},
		{
			name:   "发送验证码太频繁",
			fields: newLocalCodeCache(60, 3, 600, 1).(*LocalCodeCache),
			args: &args{
				biz:   "codeCache",
				phone: "12345678912",
				code:  "12345",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields
			if tt.name == "发送验证码太频繁" {
				c.Set(tt.args.ctx, tt.args.biz, tt.args.phone, tt.args.code)
			}
			err := c.Set(tt.args.ctx, tt.args.biz, tt.args.phone, tt.args.code)
			assert.Equal(t, tt.wantErr, err == nil)

		})
	}
}

func TestLocalCodeCache_Verify(t *testing.T) {
	type args struct {
		ctx       context.Context
		biz       string
		phone     string
		inputCode string
	}
	tests := []struct {
		name    string
		fields  *LocalCodeCache
		args    *args
		want    bool
		wantErr error
	}{
		{
			name:   "正常验证",
			fields: newLocalCodeCache(60, 3, 600, 1).(*LocalCodeCache),
			args: &args{
				biz:       "codeCache",
				phone:     "12345678901",
				inputCode: "123456",
			},
			want:    true,
			wantErr: nil,
		},
		{
			name:   "验证次数太多",
			fields: newLocalCodeCache(60, -1, 600, 1).(*LocalCodeCache),
			args: &args{
				biz:       "codeCache",
				phone:     "12345678902",
				inputCode: "123456",
			},
			want:    false,
			wantErr: ErrCodeVerifyTooManyTimes,
		},
		{
			name:   "缓存过期主动清理",
			fields: newLocalCodeCache(60, 3, 2, 60).(*LocalCodeCache),
			args: &args{
				biz:       "codeCache",
				phone:     "12345678904",
				inputCode: "123456",
			},
			want:    false,
			wantErr: ErrUnknownForCode,
		},
		{
			name:   "验证码不等",
			fields: newLocalCodeCache(60, 3, 1, 1).(*LocalCodeCache),
			args: &args{
				biz:       "codeCache",
				phone:     "12345678905",
				inputCode: "123456",
			},
			want:    false,
			wantErr: nil,
		},
		{
			name:   "缓存超时被动清理",
			fields: newLocalCodeCache(1, 3, 1, 1).(*LocalCodeCache),
			args: &args{
				biz:       "codeCache",
				phone:     "12345678906",
				inputCode: "555555",
			},
			want:    false,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields
			c.Set(tt.args.ctx, tt.args.biz, tt.args.phone, tt.args.inputCode)
			if tt.name == "缓存过期主动清理" {
				time.Sleep(time.Second * 2)
			} else if tt.name == "验证码不等" {
				tt.args.inputCode = "111111"
			} else if tt.name == "缓存超时被动清理" {
				time.Sleep(time.Second * 2)
			}
			got, err := c.Verify(tt.args.ctx, tt.args.biz, tt.args.phone, tt.args.inputCode)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
