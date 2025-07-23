package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func SetNX(ctx context.Context, client *redis.Client, key, value string, ttl time.Duration) (err error) {
	now := time.Now()
	defer func() {
		l := logrus.WithContext(ctx).WithFields(logrus.Fields{
			"start": now,
			"key":   key,
			"value": value,
			"err":   err,
			"cost":  time.Since(now).Milliseconds(),
		})
		if err == nil {
			l.Info("Redis setnx ok")
		} else {
			l.Warn("Redis setnx fail")
		}
	}()

	if client == nil {
		return errors.New("redis client is nil")
	}
	_, err = client.SetNX(ctx, key, value, ttl).Result()
	return err
}

func Del(ctx context.Context, client *redis.Client, key string) (err error) {
	now := time.Now()
	defer func() {
		l := logrus.WithContext(ctx).WithFields(logrus.Fields{
			"start": now,
			"key":   key,
			"err":   err,
			"cost":  time.Since(now).Milliseconds(),
		})
		if err == nil {
			l.Info("Redis delete ok")
		} else {
			l.Warn("Redis delete fail")
		}
	}()

	if client == nil {
		return errors.New("redis client is nil")
	}
	_, err = client.Del(ctx, key).Result()
	return err
}
