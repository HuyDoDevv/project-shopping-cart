package db

import (
	"context"
	"fmt"
	"gin/user-management-api/internal/config"
	"gin/user-management-api/internal/db/sqlc"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/pkg/loggers"
	"gin/user-management-api/pkg/pgx"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

var DB sqlc.Querier
var DBpool *pgxpool.Pool

func InitDB() error {
	conStr := config.NewConfig().DNS()

	conf, err := pgxpool.ParseConfig(conStr)

	if err != nil {
		return fmt.Errorf("error passing DB config: %v", err)
	}

	sqlLogger := utils.NewLoggerWithPath("slq.log", "info")
	conf.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger: &pgx.PgxZeroLogTracer{
			Logger:         *sqlLogger,
			SlowQueryLimit: 500 * time.Microsecond,
		},
		LogLevel: tracelog.LogLevelDebug,
	}

	conf.MaxConns = 50
	conf.MinConns = 50
	conf.MaxConnLifetime = 30 * time.Minute
	conf.MaxConnIdleTime = 5 * time.Minute
	conf.HealthCheckPeriod = 1 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	DBpool, err = pgxpool.NewWithConfig(ctx, conf)

	if err != nil {
		return fmt.Errorf("error creating DB poll: %v", err)
	}

	DB = sqlc.New(DBpool)

	if err := DBpool.Ping(ctx); err != nil {
		return fmt.Errorf("db ping error: %v", err)
	}
	loggers.Log.Info().Msg("Connected")
	return nil
}
