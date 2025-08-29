package provider

import (
	"authcenterapi/util"
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func NewPostgresConnection(ctx context.Context) (*pgx.Conn, error) {
	cfg := util.Configuration.Postgres

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		strings.Join(cfg.Options, "&"),
	)

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
