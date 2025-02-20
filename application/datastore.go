package application

import (
	"context"
	"first-little-server/order"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"log"
)

type Datastore struct {
	rdb    *redis.Client
	pgb    *pgx.Conn
	config Config
}

func NewDatastore(ctx context.Context, config Config) *Datastore {
	ds := &Datastore{
		config: config,
	}

	ds.setInnerDatabase(ctx)

	return ds
}

func (ds *Datastore) setInnerDatabase(ctx context.Context) {
	switch ds.config.Database {
	case PostgresEnv:
		postgres, err := pgx.Connect(ctx, ds.config.PostgresAddress)
		if err != nil {
			postgres = nil
		}
		ds.pgb = postgres
		ds.rdb = nil
	case ReddisEnv:
		ds.rdb = redis.NewClient(&redis.Options{
			Addr: ds.config.RedisAddress,
		})
		ds.pgb = nil
	default:
		log.Fatalf("database %s is not supported", ds.config.Database)
	}
}

// Ping the inner database to verify connexion.
func (ds *Datastore) Ping(ctx context.Context) error {
	if ds.pgb != nil {
		err := ds.pgb.Ping(ctx)

		if err != nil {
			return fmt.Errorf("error when pinging postgres: %w", err)
		}
		return nil
	}

	if ds.rdb != nil {
		err := ds.rdb.Ping(ctx).Err()

		if err != nil {
			return fmt.Errorf("error when pinging reddis: %w", err)
		}
		return nil
	}

	return fmt.Errorf("database %s not supported", ds.config.Database)
}

// Close the inner database.
func (ds *Datastore) Close(ctx context.Context) error {
	if ds.pgb != nil {
		if err := ds.pgb.Close(ctx); err != nil {
			return fmt.Errorf("failed to close postgres: %w", err)
		}
		return nil
	}

	if ds.rdb != nil {
		if err := ds.rdb.Close(); err != nil {
			return fmt.Errorf("failed to close reddis: %w", err)
		}
		return nil
	}

	return fmt.Errorf("database %s not supported", ds.config.Database)
}

// GetActiveRepo returns the current active repository
// If no current repository is active, returns null.
func (ds *Datastore) GetActiveRepo() order.Repository {
	if ds.pgb != nil {
		return &order.PostgresRepo{
			Client: ds.pgb,
		}
	}

	if ds.rdb != nil {
		return &order.RedisRepo{
			Client: ds.rdb,
		}
	}

	return nil
}
