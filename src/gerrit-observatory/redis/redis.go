package redis

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"strings"
	"time"
)

var (
	redisPool          *redis.Pool
	subscribeKeyPrefix = "subscribe:"
	idKey              = "subscribe_id_index"
	redisMaxIdle       = 3
	redisIdleTimeout   = 240 * time.Second
)

// Subscribe ...
type Subscribe struct {
	ID               int
	Detail           SubscribeDetail
	CreatedTime      string
	LastActivateTime string
	ActivateCount    int
	Valid            bool
}

func (sub *Subscribe) Activate() (err error) {
	key := fmt.Sprintf("%s%d", subscribeKeyPrefix, sub.ID)

	redisConn := redisPool.Get()
	defer redisConn.Close()

	exist, err := redis.Bool(redisConn.Do("EXISTS", key))
	if err != nil {
		return
	}
	if !exist {
		return fmt.Errorf("subscribe %s not existed", key)
	}
	now := time.Now().Format(time.UnixDate)
	_, err = redisConn.Do("HSET", key, "last_active_time", now)
	if err != nil {
		return
	}
	_, err = redisConn.Do("HINCRBY", key, "active_count", 1)
	if err != nil {
		return
	}
	sub.LastActivateTime = now
	sub.ActivateCount++
	return
}

func (sub *Subscribe) Invalid() (err error) {
	redisConn := redisPool.Get()
	defer redisConn.Close()

	key := fmt.Sprintf("%s%d", subscribeKeyPrefix, sub.ID)
	_, err = redisConn.Do("HSET", key, "valid", false)
	if err != nil {
		return
	}
	sub.Valid = false
	return
}

// SubscribeDetail ...
type SubscribeDetail struct {
	Filter  map[string]interface{} `json:"filter"`
	HookURL string                 `json:"hook_url"`
	Comment string                 `json:"comment"`
}

func (subd *SubscribeDetail) Save() (id int, err error) {
	redisConn := redisPool.Get()
	defer redisConn.Close()

	id, err = redis.Int(redisConn.Do("INCR", idKey))
	if err != nil {
		return
	}
	key := fmt.Sprintf("%s%d", subscribeKeyPrefix, id)
	createdTime := time.Now().Format(time.UnixDate)
	rawDetail, err := json.Marshal(*subd)
	if err != nil {
		return
	}
	ok, err := redis.String(
		redisConn.Do("HMSET", key,
			"detail", rawDetail,
			"created_time", createdTime,
			"activate_count", 0,
			"last_activate_time", "",
			"valid", true))
	if err != nil {
		return
	}
	if ok != "OK" {
		err = fmt.Errorf("hmset failed, %s", ok)
		return
	}
	return
}

func GetSubscribes() ([]*Subscribe, error) {
	subscribes := make([]*Subscribe, 0, 10)

	redisConn := redisPool.Get()
	defer redisConn.Close()

	keys, err := redis.ByteSlices(redisConn.Do("KEYS", subscribeKeyPrefix+"*"))
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		subscribe, err := getSubscribeByKey(string(key))
		if err != nil {
			return nil, err
		}
		subscribes = append(subscribes, subscribe)
	}
	return subscribes, nil
}

func GetSubscribe(id int) (*Subscribe, error) {
	key := getKey(id)
	return getSubscribeByKey(key)
}

func getSubscribeByKey(key string) (*Subscribe, error) {
	var (
		detail           SubscribeDetail
		createdTime      string
		lastActivateTime string
		activateCount    int
		valid            bool
	)

	redisConn := redisPool.Get()
	defer redisConn.Close()

	reply, err := redis.Values(redisConn.Do("HMGET", key, "created_time", "last_active_time", "active_count", "valid"))
	if err != nil {
		return nil, err
	}

	_, err = redis.Scan(reply, &createdTime, &lastActivateTime, &activateCount, &valid)
	if err != nil {
		return nil, err
	}

	detailRaw, err := redis.Bytes(redisConn.Do("HGET", key, "detail"))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(detailRaw, &detail)
	if err != nil {
		return nil, err
	}

	id, err := strconv.Atoi(strings.TrimPrefix(key, subscribeKeyPrefix))
	if err != nil {
		return nil, err
	}

	return &Subscribe{
		ID:               id,
		Detail:           detail,
		CreatedTime:      createdTime,
		LastActivateTime: lastActivateTime,
		ActivateCount:    activateCount,
		Valid:            valid,
	}, nil
}

func DeleteSubscribe(id int) (bool, error) {
	key := getKey(id)
	return deleteSubscribeByKey(key)
}

func deleteSubscribeByKey(key string) (bool, error) {
	redisConn := redisPool.Get()
	defer redisConn.Close()

	b, err := redis.Bool(redisConn.Do("DEL", key))
	if err != nil {
		return false, err
	} else {
		return b, nil
	}
}

func getKey(id int) string {
	return fmt.Sprintf("%s%d", subscribeKeyPrefix, id)
}


func InitRedis(host string, port int, db int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	dbOption := redis.DialDatabase(db)
	redisPool = newRedisPool(addr, dbOption)
}

func DestroyRedis() {
	redisPool.Close()
}

func newRedisPool(addr string, options ...redis.DialOption) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     redisMaxIdle,
		IdleTimeout: redisIdleTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr, options...)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
