package localcache

import (
	"context"
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/repository/cache"
	"strings"
	"sync"
	"time"
)

var (
	ErrCodeSendTooMany        = cache.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
	ErrUnknownForCode         = cache.ErrUnknownForCode
)

type CodeCache interface {
	Set(ctx context.Context, biz string,
		phone string, code string) error

	Verify(ctx context.Context, biz string,
		phone string, inputCode string) (bool, error)
}

// CodeCache 基于本地内存的实现
type LocalCodeCache struct {
	lock       sync.Mutex
	m          map[string]string
	ttl        map[string]time.Time
	cnt        map[string]int
	expiretime int  //缓存过期时间（秒）
	timeout    int  //校验码有效时间（秒）
	times      int  //校验码有效次数
	first      bool //标记首次运行
	interval   int  //定时扫描清理缓存的时间间隔(秒)
}

// 无参数构造，方便wire构建
func NewCodeCache() CodeCache {
	return newLocalCodeCache(60, 3, 600, 1)
}

// 带参数构造，方便测试
func newLocalCodeCache(timeout, times, expiretime, interval int) CodeCache {
	return &LocalCodeCache{
		lock:       sync.Mutex{},
		m:          make(map[string]string, 1),
		ttl:        make(map[string]time.Time),
		cnt:        make(map[string]int),
		timeout:    timeout,
		times:      times,
		expiretime: expiretime,
		first:      true,
		interval:   interval,
	}
}

func (c *LocalCodeCache) Set(ctx context.Context,
	biz string,
	phone string,
	code string) error {

	c.lock.Lock()
	defer c.lock.Unlock()

	if c.first {
		go c.cleanup()
		c.first = false
	}

	ttlkey := c.makettlkey(biz, phone)

	ttl, ok := c.ttl[ttlkey]
	if ok {
		end := time.Now()
		duration := end.Sub(ttl)
		//短信发送时间间隔少于timeout
		if int(duration.Seconds()) < c.timeout {
			return ErrCodeSendTooMany
		}
	}
	//没有key或发送超时重新发送
	key := c.makekey(biz, phone)
	cntkey := c.makecntkey(biz, phone)
	c.m[key] = code
	c.cnt[cntkey] = c.times
	c.ttl[ttlkey] = time.Now()
	return nil
}

func (c *LocalCodeCache) Verify(ctx context.Context,
	biz string, phone string, inputCode string) (bool, error) {

	c.lock.Lock()
	defer c.lock.Unlock()
	cntkey := c.makecntkey(biz, phone)
	key := c.makekey(biz, phone)
	ttlkey := c.makettlkey(biz, phone)
	ttl := c.ttl[ttlkey]
	cnt, ok := c.cnt[cntkey]
	if ok {
		if cnt <= 0 {
			return false, ErrCodeVerifyTooManyTimes
		} else {
			code, ok := c.m[key]
			if !ok {
				return false, ErrUnknownForCode
			}
			//如果key存在，要判断是否超时
			end := time.Now()
			duration := end.Sub(ttl)
			//超时
			if int(duration.Seconds()) >= c.expiretime {
				//清理相关的key的缓存
				c.cleanCache(biz, phone)
				return false, ErrUnknownForCode
			}
			if code == inputCode {
				c.cleanCache(biz, phone)
				return true, nil
			} else {
				c.cnt[cntkey] = c.cnt[cntkey] - 1
			}
		}
	}
	//找不到key或者验证失败
	return false, nil
}

func (c *LocalCodeCache) cleanCache(biz, phone string) {
	delete(c.m, c.makekey(biz, phone))
	delete(c.ttl, c.makettlkey(biz, phone))
	delete(c.cnt, c.makecntkey(biz, phone))
}

// 定时清理过期缓存,验证码发出去了但一直不来验证，无法主动清理，因此需要此方法
func (c *LocalCodeCache) cleanup() {
	for {
		time.Sleep(time.Duration(c.interval) * time.Second) // 每隔一段时间执行一次清理操作

		func() {
			c.lock.Lock()
			defer c.lock.Unlock()
			for k, ttl := range c.ttl {
				end := time.Now()
				if int(end.Sub(ttl).Seconds()) > c.expiretime {
					biz, phone := c.getItmefromKey(k)
					c.cleanCache(biz, phone)
				}
			}
		}()

	}
}
func (c *LocalCodeCache) makekey(biz string,
	phone string) string {
	return fmt.Sprintf("%s:%s", biz, phone)
}

func (c *LocalCodeCache) makettlkey(biz string,
	phone string) string {
	return fmt.Sprintf("%s:%s:ttl", biz, phone)
}

func (c *LocalCodeCache) makecntkey(biz string,
	phone string) string {
	return fmt.Sprintf("%s:%s:cnt", biz, phone)
}

func (c *LocalCodeCache) getItmefromKey(key string) (biz, code string) {
	s := strings.Split(key, ":")
	return s[0], s[1]
}
