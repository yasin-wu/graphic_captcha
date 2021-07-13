package redis

import (
	"strings"
	"time"

	"log"

	"github.com/gomodule/redigo/redis"
)

var (
	RedisPool *redis.Pool
)

type RedisConfig struct {
	Host           string //Redis地址
	Db             int    //使用的数据库
	PassWord       string //密码
	NetWork        string //网络协议
	MaxIdle        int    //连接池最大空闲连接数
	MaxActive      int    //连接池最大激活连接数
	ConnectTimeout int    //连接超时,单位毫秒
	ReadTimeout    int    //读取超时,单位毫秒
	WriteTimeout   int    //写入超时,单位毫秒
	ExpireTime     int    //key过期时间,单位秒
}

func InitRedisPool(conf *RedisConfig) {
	if conf == nil {
		panic("redis config is nil")
	}
	redisDial := func() (redis.Conn, error) {
		conn, err := redis.Dial(
			strings.ToLower(conf.NetWork),
			conf.Host,
			redis.DialConnectTimeout(time.Duration(conf.ConnectTimeout)*time.Millisecond),
			redis.DialReadTimeout(time.Duration(conf.ReadTimeout)*time.Millisecond),
			redis.DialWriteTimeout(time.Duration(conf.WriteTimeout)*time.Millisecond),
		)
		if err != nil {
			log.Printf("连接redis失败:%s", err.Error())
			return nil, err
		}

		if conf.PassWord != "" {
			if _, err := conn.Do("AUTH", conf.PassWord); err != nil {
				conn.Close()
				log.Printf("redis认证失败:%s", err.Error())
				return nil, err
			}
		}

		_, err = conn.Do("SELECT", conf.Db)
		if err != nil {
			conn.Close()
			log.Printf("redis选择数据库失败:%s", err.Error())
			return nil, err
		}

		return conn, nil
	}

	redisTestOnBorrow := func(conn redis.Conn, t time.Time) error {
		_, err := conn.Do("PING")
		if err != nil {
			log.Printf("从redis连接池取出的连接无效:%s", err.Error())
		}
		return err
	}

	RedisPool = &redis.Pool{
		MaxIdle:      conf.MaxIdle,
		MaxActive:    conf.MaxActive,
		IdleTimeout:  300 * time.Second,
		Dial:         redisDial,
		TestOnBorrow: redisTestOnBorrow,
		Wait:         true,
	}
}

func ExecRedisCommand(command string, args ...interface{}) (interface{}, error) {
	conn := RedisPool.Get()
	defer conn.Close()
	return conn.Do(command, args...)
}
