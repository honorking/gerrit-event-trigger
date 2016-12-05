package redis

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

var (
	DetailRaw = `{
	  "filter": {
		"type": "patchset-created",
		"patchSet": {
		  "uploader": {
			"email": "guoxuxing@wandoulabs.com"
		  }
		},
		"change": {
		  "branch": "release-*",
		  "project": "loki"
		}
	  },
	  "hook_url": "http://loki.wandoulabs.com/webhook?from_gerrit", 
	  "comment": "用途说明"
	}`
)

type RedisTestSuite struct {
	suite.Suite
	redisConn redis.Conn
}

func (suite *RedisTestSuite) SetupSuite() {
	InitRedis("127.0.0.1", 6379, 1)
}

func (suite *RedisTestSuite) TearDownSuite() {
	DestroyRedis()
}

func (suite *RedisTestSuite) SetupTest() {
	suite.redisConn = redisPool.Get()
	suite.redisConn.Do("FLUSHDB")
}

func (suite *RedisTestSuite) TearDownTest() {
	suite.redisConn.Close()
	suite.redisConn = nil
}

func (suite *RedisTestSuite) TestSubscribeDetailSaveGetActivate() {
	detail := SubscribeDetail{}
	err := json.Unmarshal([]byte(DetailRaw), &detail)
	assert.Nil(suite.T(), err)
	id, err := detail.Save()
	assert.Nil(suite.T(), err)
	assert.NotEqual(suite.T(), id, 0, "id should not be zero")
	subscribe, err := GetSubscribe(id)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), subscribe.ActivateCount, 0, "ActiveCount should be 0")
	assert.Equal(suite.T(), subscribe.LastActivateTime, "", "lastactiveTime should be empty")
	subscribe.Activate()
	assert.Equal(suite.T(), subscribe.ActivateCount, 1, "ActiveCount should be 1")
	assert.NotEqual(suite.T(), subscribe.LastActivateTime, "", "lastactiveTime should not be empty")
	newSubscirbe, err := GetSubscribe(id)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), subscribe.ActivateCount, newSubscirbe.ActivateCount, "fail update redis activeCount")
	assert.Equal(suite.T(), subscribe.LastActivateTime, newSubscirbe.LastActivateTime, "fail update redis lastActivateTime")
}

func (suite *RedisTestSuite) TestGetSubscribes() {
	detail := SubscribeDetail{}
	err := json.Unmarshal([]byte(DetailRaw), &detail)
	assert.Nil(suite.T(), err)
	id, err := detail.Save()
	assert.Nil(suite.T(), err)
	subscribes, err := GetSubscribes()
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), subscribes, 1, "subscribes length should be 1")
	assert.Equal(suite.T(), subscribes[0].ID, id, "id not meet")
}

func (suite *RedisTestSuite) TestSubscribeInvalid() {
	detail := SubscribeDetail{}
	err := json.Unmarshal([]byte(DetailRaw), &detail)
	assert.Nil(suite.T(), err)
	id, err := detail.Save()
	assert.Nil(suite.T(), err)
	subscribe, err := GetSubscribe(id)
	err = subscribe.Invalid()
	assert.Nil(suite.T(), err)
	newSubscirbe, err := GetSubscribe(id)
	assert.Equal(suite.T(), subscribe.Valid, newSubscirbe.Valid, "fail update redis validity")
	assert.Equal(suite.T(), newSubscirbe.Valid, false)
}

func TestRedisTestSuite(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}
