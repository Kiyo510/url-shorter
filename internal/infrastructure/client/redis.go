package client

import (
	"github.com/Kiyo510/url-shorter/internal/config"
	"github.com/redis/rueidis"
	"log"
	"sync"
)

var (
	client rueidis.Client
	once   sync.Once
)

// GetRedisClient はシングルトンパターンを使用して Redis クライアントを生成・取得します。
func GetRedisClient() rueidis.Client {
	conf := config.RedisConf
	once.Do(func() {
		var err error
		client, err = rueidis.NewClient(rueidis.ClientOption{
			InitAddress: []string{conf.Host + ":" + conf.Port},
		})
		if err != nil {
			log.Fatalf("Failed to connect to Redis: %v", err)
		}
	})
	return client
}
