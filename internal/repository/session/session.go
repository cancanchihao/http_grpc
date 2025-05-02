package session

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"http_grpc/pkg/pool"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

// SessionStore 结构体
type SessionStore struct {
	ID         string
	LastAccess time.Time
	Values     map[string]interface{}
}

// Provider 结构体：管理所有session
type Provider struct {
	sessions    map[string]*list.Element // sessionID -> SessionStore
	list        *list.List               // 便于 GC 管理
	lock        sync.RWMutex             // 读写锁
	maxLifeTime time.Duration            // 超时时间
}

// 创建一个全局 Provider
var provider = &Provider{
	sessions: make(map[string]*list.Element),
	list:     list.New(),
}

// Redis 客户端
var rdb *redis.Client
var ctx = context.Background()

const sessionRedisPrefix = "session:"

// InitRedis 初始化 Redis
func InitRedis(addr, password string, db int) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	return err
}

// LoadSessionsFromRedis 启动时从 Redis 恢复 Session
func LoadSessionsFromRedis() error {
	keys, err := rdb.Keys(ctx, sessionRedisPrefix+"*").Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		// 取回 Session 数据
		data, getSessionErr := rdb.HGetAll(ctx, key).Result()
		if getSessionErr != nil || len(data) == 0 {
			continue
		}

		// 恢复数据格式
		lastAccessUnix, _ := strconv.ParseInt(data["last_access"], 10, 64)
		// 过滤超时信息
		if lastAccess := time.Unix(0, lastAccessUnix); lastAccess.Add(provider.maxLifeTime).Before(time.Now()) {
			// 已过期，删除
			rdb.Del(ctx, key)
			continue
		}

		var values map[string]interface{}
		unmarshalErr := json.Unmarshal([]byte(data["values"]), &values)
		if unmarshalErr != nil {
			return unmarshalErr
		}

		store := &SessionStore{
			ID:         data["id"],
			LastAccess: time.Unix(0, lastAccessUnix),
			Values:     make(map[string]interface{}),
		}

		// 这里的 Values 简化处理：可以根据需要扩展
		provider.lock.Lock()
		element := provider.list.PushFront(store)
		provider.sessions[store.ID] = element
		provider.lock.Unlock()
	}
	return nil
}

// GetSession 获取或者创建新的 SessionStore
func GetSession(c *gin.Context) *SessionStore {
	var store *SessionStore

	// 获取session_id cookie
	cookie, err := c.Cookie("session_id")
	if err == nil {
		sessionID := cookie
		provider.lock.RLock()
		if element, ok := provider.sessions[sessionID]; ok {
			store = element.Value.(*SessionStore)
		}
		provider.lock.RUnlock()
	}

	if store == nil {
		// 创建新 Session
		store = newSession()
		provider.lock.Lock()
		element := provider.list.PushFront(store) // 新的放到链表头
		provider.sessions[store.ID] = element
		provider.lock.Unlock()

		// 设置到 Cookie
		c.SetCookie("session_id", store.ID, 30*60, "/", "", false, true)
	} else {
		// 更新最后访问时间并移动到链表头
		provider.lock.Lock()
		store.LastAccess = time.Now()
		provider.list.MoveToFront(provider.sessions[store.ID])
		provider.lock.Unlock()
	}

	// 更新 Redis
	saveSessionToRedis(store)

	return store
}

// 创建新的 SessionStore
func newSession() *SessionStore {
	return &SessionStore{
		ID:         generateSessionID(),
		LastAccess: time.Now(),
		Values:     make(map[string]interface{}),
	}
}

// 保存 SessionStore 到 Redis
func saveSessionToRedis(store *SessionStore) {
	key := sessionRedisPrefix + store.ID

	valueBytes, err := json.Marshal(store.Values)
	if err != nil {
		// 打印日志，不阻止流程
		fmt.Println("Failed to marshal session values:", err)
		valueBytes = []byte("{}")
	}

	pool.SessionPool.AddTask(pool.Task{
		Job: func() error {
			rdb.HSet(ctx, key, map[string]interface{}{
				"id":          store.ID,
				"last_access": store.LastAccess.UnixNano(),
				"values":      string(valueBytes),
			})
			rdb.Expire(ctx, key, 120*time.Minute)
			return nil
		},
	})
}

// 删除 SessionStore 从 Redis
func deleteSessionFromRedis(sessionID string) {
	key := sessionRedisPrefix + sessionID
	pool.SessionPool.AddTask(pool.Task{
		Job: func() error {
			rdb.Del(ctx, key)
			return nil
		},
	})
}

// SessionGC 清理过期的 Session
func SessionGC() {
	provider.lock.Lock()
	defer provider.lock.Unlock()

	for {
		element := provider.list.Back() // 从尾部开始（最久未访问的）
		if element == nil {
			break
		}

		// 如果已经过期
		if store := element.Value.(*SessionStore); store.LastAccess.Add(provider.maxLifeTime).Before(time.Now()) {
			// 从 list 和 map 中移除
			provider.list.Remove(element)
			delete(provider.sessions, store.ID)
			pool.SessionPool.AddTask(pool.Task{
				Job: func() error {
					deleteSessionFromRedis(store.ID)
					return nil
				},
			})
		} else {
			// list 按时间顺序，后面都不会过期
			break
		}
	}
}

// StartGC 启动后台Session回收协程
func StartGC(maxLifetime time.Duration) {
	ticker := time.NewTicker(1 * time.Minute) // 每1分钟检查一次
	provider.maxLifeTime = maxLifetime
	go func() {
		for {
			<-ticker.C
			SessionGC()
		}
	}()
}

// 生成简单的随机 SessionID
func generateSessionID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + strconv.Itoa(rand.Intn(100000))
}
