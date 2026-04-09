package provider

import (
	"authcenterapi/util"
	"context"

	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

func NewRedisConnection(ctx context.Context) (*redis.Client, error) {
	cfg := util.Configuration.Redis // pastikan struct Redis ada di config

	// Format Redis URL: redis://<username>:<password>@<host>:<port>/<db>?<options>
	// Username biasanya optional, Redis default tidak pakai user.
	dsn := fmt.Sprintf(
		"redis://%s:%s@%s:%d/%d?%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DB,
		strings.Join(cfg.Options, "&"),
	)

	opt, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}

	client := redis.NewClient(opt)

	// Test koneksi
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
}
