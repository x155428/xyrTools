package fileMonitor

import (
	"sync"
	"time"
)

// 行为缓存配置
type BehaviorCacheConfig struct {
	Window     time.Duration
	Threshold  int
	Resolution time.Duration
}

type pathStat struct {
	Timestamps []int64
}

type BehaviorCache struct {
	mu       sync.Mutex
	cache    map[string]*pathStat
	config   BehaviorCacheConfig
	nowFunc  func() time.Time
	cleanGap time.Duration
}

func NewBehaviorCache(cfg BehaviorCacheConfig) *BehaviorCache {
	if cfg.Resolution <= 0 {
		cfg.Resolution = time.Second
	}
	return &BehaviorCache{
		cache:    make(map[string]*pathStat),
		config:   cfg,
		nowFunc:  time.Now,
		cleanGap: cfg.Window * 2,
	}
}

func (bc *BehaviorCache) Record(path string) bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	now := bc.nowFunc().Unix()
	key := truncateTime(now, bc.config.Resolution)

	stat, ok := bc.cache[path]
	if !ok {
		stat = &pathStat{}
		bc.cache[path] = stat
	}
	stat.Timestamps = append(stat.Timestamps, key)

	cutoff := key - int64(bc.config.Window.Seconds())
	var filtered []int64
	for _, t := range stat.Timestamps {
		if t >= cutoff {
			filtered = append(filtered, t)
		}
	}
	stat.Timestamps = filtered

	return len(stat.Timestamps) >= bc.config.Threshold
}

func truncateTime(ts int64, res time.Duration) int64 {
	return ts - (ts % int64(res.Seconds()))
}
