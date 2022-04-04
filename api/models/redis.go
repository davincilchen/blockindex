package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

var redisConn *RedisClient

//var redisClient *redis.Client

type RedisConfig struct {
	Addr string `env:"REDIS_ADDR" envDefault:"127.0.0.1"`
	Port string `env:"REDIS_PORT" envDefault:"6379"`
}

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     20,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute { //duration
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}
func InitRedis(cfg *RedisConfig) {
	addr := fmt.Sprintf("%s:%s", cfg.Addr, cfg.Port)
	p := newPool(addr)
	redisConn = &RedisClient{
		Pool: p,
	}

}



// ============== //
var ErrRedisConnNil = errors.New("RedisConn is nil")

//args: [key,value]
func RedisSetWithExpire(expireSec int64, args ...interface{}) (interface{}, error) { // reply , err

	if redisConn == nil {
		return nil, ErrRedisConnNil
	}

	return redisConn.SetWithExpire(expireSec, args...)
}

func RedisSet(args ...interface{}) (interface{}, error) { // reply , err

	if redisConn == nil {
		return nil, ErrRedisConnNil
	}

	return redisConn.Set(args...)
}

func RedisGet(key string) (interface{}, error) { // reply , err

	if redisConn == nil {
		return nil, ErrRedisConnNil
	}

	return redisConn.Get(key)
}

type RedisClient struct {
	Pool *redis.Pool
}

func (r *RedisClient) Delete(key string) (interface{}, error) { // reply , err
	conn := r.Pool.Get()
	defer conn.Close()

	rp, err := conn.Do("Del", key)
	if err != nil {
		return rp, fmt.Errorf("RedisClient Delete error key [%v]: %v", key, err)
	}

	return rp, nil

}

//args: [key,value]
func (r *RedisClient) SetWithExpire(expireSec int64, args ...interface{}) (interface{}, error) { // reply , err
	conn := r.Pool.Get()
	defer conn.Close()

	l := len(args)
	if l <= 0 {
		return nil, fmt.Errorf("RedisClient SetWithExpire len(args) <=0 ")
	}
	key := args[0]
	totalLen := l + 2
	newArgs := make([]interface{}, totalLen)
	copy(newArgs, args)
	newArgs[2] = "EX"
	newArgs[3] = expireSec
	rp, err := conn.Do("Set", newArgs...)
	if err != nil {
		return rp, fmt.Errorf("RedisClient SetWithExpire Set error key [%v]: %v", key, err)
	}
	return rp, nil

}

func (r *RedisClient) Set(args ...interface{}) (interface{}, error) { // reply , err
	conn := r.Pool.Get()
	defer conn.Close()
	if len(args) <= 0 {
		return nil, fmt.Errorf("RedisClient Set len(args) <=0 ")
	}
	key := args[0]
	rp, err := conn.Do("Set", args...)
	if err != nil {
		return rp, fmt.Errorf("RedisClient Set error key [%v]: %v", key, err)
	}

	return rp, nil

}

func (r *RedisClient) Exp(key string, timeint int64) (int64, error) {
	conn := r.Pool.Get()
	defer conn.Close()
	data, err := redis.Int64(conn.Do("EXPIRE", key, timeint))

	if err != nil {
		return data, fmt.Errorf("RedisClient Exp error key %s: %v", key, err)
	}
	return data, nil
}

func (r *RedisClient) Pexp(key string, timeint int64) (int64, error) {
	conn := r.Pool.Get()
	defer conn.Close()
	data, err := redis.Int64(conn.Do("PEXPIRE", key, timeint))

	if err != nil {
		return data, fmt.Errorf("RedisClient Pexp error key %s: %v", key, err)
	}
	return data, nil
}

func (r *RedisClient) Get(key string) (interface{}, error) { // reply , err
	conn := r.Pool.Get()
	defer conn.Close()
	rp, err := conn.Do("Get", key)
	if err != nil {
		return rp, fmt.Errorf("RedisClient Get error key [%v]: %v", key, err)
	}

	return rp, nil
}
