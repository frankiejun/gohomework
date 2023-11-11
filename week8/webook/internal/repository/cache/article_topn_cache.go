package cache

import (
	"container/heap"
	"context"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
	"sync"
	"time"
)

type ArticleTopnCache interface {
	Add(ctx context.Context, artId int64, cnt int) error
	Get(ctx context.Context, n int) ([]domain.ArticlTopN, error)
	Flush(ctx context.Context) error
}

type ArticleInfo struct {
	artId   int64
	likeCnt int64
}

type articleHeap []ArticleInfo

func (h articleHeap) Len() int { return len(h) }
func (h articleHeap) Less(i, j int) bool {
	return h[i].likeCnt > h[j].likeCnt
}
func (h articleHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
func (h *articleHeap) Push(x any) {
	*h = append(*h, x.(ArticleInfo))
}

func (h *articleHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type ArticleTopnCacheImpl struct {
	client  redis.Cmdable
	artdata articleHeap
	num     int
	locker  sync.Mutex
	dataCnt map[int64]int
}

func NewArticleTopnCacheImpl(client redis.Cmdable) ArticleTopnCache {
	return newArticleTopnCache(client, 100, 10)
}

func newArticleTopnCache(client redis.Cmdable, num, interval int) ArticleTopnCache {
	ret := &ArticleTopnCacheImpl{
		client:  client,
		artdata: make(articleHeap, 0),
		num:     num,
		dataCnt: make(map[int64]int),
		locker:  sync.Mutex{},
	}
	//定时自动刷新缓存到redis
	go func() {
		for {
			ret.Flush(context.Background())
			time.Sleep(time.Second * time.Duration(interval))
		}
	}()
	return ret
}

// 把当前小堆中的点赞数提交到redis
func (a *ArticleTopnCacheImpl) Flush(ctx context.Context) error {
	a.locker.Lock()
	defer a.locker.Unlock()

	//建小堆
	for artid, cnt := range a.dataCnt {
		heap.Push(&a.artdata, ArticleInfo{
			artId:   artid,
			likeCnt: int64(cnt)})
		if len(a.artdata) > a.num {
			heap.Pop(&a.artdata)
		}
		a.dataCnt[artid] = 0 //点赞数缓存清0
	}
	//把构建好的堆推到redis,和redis上原有的数据进行累加
	for _, art := range a.artdata {
		a.client.ZIncrBy(ctx, "articleTopLikes", float64(art.likeCnt), strconv.Itoa(int(art.artId)))
	}
	a.artdata = a.artdata[:0] //清空堆

	return nil
}

// 加到本地缓存计数
func (a *ArticleTopnCacheImpl) Add(ctx context.Context, artId int64, cnt int) error {
	a.locker.Lock()
	defer a.locker.Unlock()

	if _, ok := a.dataCnt[artId]; ok {
		a.dataCnt[artId] += cnt
	} else {
		a.dataCnt[artId] = cnt
	}
	return nil
}

func toEntity(z redis.Z) domain.ArticlTopN {
	artsid, _ := strconv.Atoi(z.Member.(string))
	return domain.ArticlTopN{
		Id:      int64(artsid),
		LikeCnt: int64(z.Score),
	}
}

// 从redis中取top n点赞数的文章id
func (a *ArticleTopnCacheImpl) Get(ctx context.Context, n int) ([]domain.ArticlTopN, error) {
	if n > a.num {
		n = a.num
	}
	a.Flush(ctx)
	val, err := a.client.ZRevRangeWithScores(ctx, "articleTopLikes", 0, int64(n)-1).Result()
	if err != nil {
		return nil, err
	}
	ret := make([]domain.ArticlTopN, len(val))
	for i, z := range val {
		ret[i] = toEntity(z)
	}
	return ret, nil
}
