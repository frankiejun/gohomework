package cache

import (
	"context"
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestArticleTopnCacheImpl_Add(t *testing.T) {
	type fields struct {
		client  redis.Cmdable
		artdata articleHeap
		num     int
		locker  sync.Mutex
		dataCnt map[int64]int
	}
	type args struct {
		ctx   context.Context
		artId int64
		cnt   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ArticleTopnCacheImpl{
				client:  tt.fields.client,
				artdata: tt.fields.artdata,
				num:     tt.fields.num,
				locker:  tt.fields.locker,
				dataCnt: tt.fields.dataCnt,
			}
			tt.wantErr(t, a.Add(tt.args.ctx, tt.args.artId, tt.args.cnt), fmt.Sprintf("Add(%v, %v, %v)", tt.args.ctx, tt.args.artId, tt.args.cnt))
		})
	}
}

func TestArticleTopnCacheImpl_All(t *testing.T) {
	redisData := []domain.ArticlTopN{
		domain.ArticlTopN{
			Id:      1,
			LikeCnt: 2,
		},
		domain.ArticlTopN{
			Id:      2,
			LikeCnt: 60,
		},
		domain.ArticlTopN{
			Id:      3,
			LikeCnt: 3,
		},
		domain.ArticlTopN{
			Id:      4,
			LikeCnt: 89,
		},
		domain.ArticlTopN{
			Id:      5,
			LikeCnt: 90,
		},
	}
	redis_client := InitRedis()
	defer redis_client.Del(context.Background(), "articleTopLikes")
	for _, v := range redisData {
		redis_client.ZIncrBy(context.Background(), "articleTopLikes", float64(v.LikeCnt), fmt.Sprintf("%d", v.Id))
	}
	type args struct {
		ctx   context.Context
		artId int64
		cnt   int
	}

	tests := []struct {
		name string
		args args
		num  int
		want []domain.ArticlTopN
	}{
		{
			name: "添加数据",
			args: args{
				ctx:   context.Background(),
				artId: 5,
				cnt:   1,
			},
			num: 3,
			want: []domain.ArticlTopN{
				domain.ArticlTopN{
					Id:      5,
					LikeCnt: 91,
				},
				domain.ArticlTopN{
					Id:      4,
					LikeCnt: 89,
				},
				domain.ArticlTopN{
					Id:      2,
					LikeCnt: 60,
				},
			},
		},
		{
			name: "添加数据,改变原有排名",
			args: args{
				ctx:   context.Background(),
				artId: 1,
				cnt:   200,
			},
			num: 3,
			want: []domain.ArticlTopN{
				domain.ArticlTopN{
					Id:      1,
					LikeCnt: 202,
				},
				domain.ArticlTopN{
					Id:      5,
					LikeCnt: 91,
				},
				domain.ArticlTopN{
					Id:      4,
					LikeCnt: 89,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewArticleTopnCacheImpl(redis_client)
			a.Add(tt.args.ctx, tt.args.artId, tt.args.cnt)
			//time.Sleep(time.Second * 2)
			got, err := a.Get(tt.args.ctx, tt.num)
			assert.NoError(t, err)
			fmt.Printf("got:%v", got)
			assert.Equal(t, tt.want, got)
		})
	}

}

func InitRedis() redis.Cmdable {
	// 这里演示读取特定的某个字段
	cmd := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return cmd
}
