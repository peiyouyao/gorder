package redis

import (
	"strconv"
	"time"

	"github.com/peiyouyao/gorder/common/handler/factory"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

const (
	confName      = "redis"
	localSupplier = "local"
)

var singleton = factory.NewSingleton(supplier)

var supplier func(string) any = func(key string) any {
	confKey := confName + "." + key
	type Section struct {
		IP           string        `mapstructure:"ip"`
		Port         int           `mapstructure:"port"`
		PoolSize     int           `mapstructure:"pool-size"`
		MaxConn      int           `mapstructure:"max-conn"`
		ConnTimeout  time.Duration `mapstructure:"conn-timeout"`
		ReadTimeout  time.Duration `mapstructure:"read-timeout"`
		WriteTimeout time.Duration `mapstructure:"write-timeout"`
	}
	var c Section
	if err := viper.UnmarshalKey(confKey, &c); err != nil {
		panic(err)
	}
	return redis.NewClient(&redis.Options{
		Network:         "tcp",
		Addr:            c.IP + ":" + strconv.Itoa(c.Port),
		PoolSize:        c.PoolSize,
		MaxRetries:      c.MaxConn,
		ConnMaxLifetime: c.ConnTimeout * time.Millisecond,
		ReadTimeout:     c.ReadTimeout * time.Millisecond,
		WriteTimeout:    c.WriteTimeout * time.Millisecond,
	})
}

func Init() {
	conf := viper.GetStringMap(confName)
	for supplyName := range conf {
		Client(supplyName)
	}
}

func Client(name string) *redis.Client {
	return singleton.Get(name).(*redis.Client)
}

func LocaClient() *redis.Client {
	return Client(localSupplier)
}
