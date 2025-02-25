package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func orderIdKey(id int64) string {
	return fmt.Sprintf("order:%d", id)
}

// Insert an order in the redis database.
func (repo *RedisRepo) Insert(ctx context.Context, order Order) error {
	data, err := json.Marshal(order)

	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}

	key := orderIdKey(order.OrderID)

	txPipe := repo.Client.TxPipeline()

	// Set overwrites data when it exists already, thus the usage of SetNX instead.
	res := txPipe.SetNX(ctx, key, string(data), 0)

	if res.Err() != nil {
		txPipe.Discard()
		return fmt.Errorf("failed to set: %w", res.Err())
	}

	if err := txPipe.SAdd(ctx, "orders", key).Err(); err != nil {
		txPipe.Discard()
		return fmt.Errorf("failed to add to orders set: %w", err)
	}

	if _, err := txPipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec insertion: %w", err)
	}

	return nil
}

var ErrNotExist = errors.New("order does not exist")

func (repo *RedisRepo) FindByID(ctx context.Context, id int64) (Order, error) {
	key := orderIdKey(id)

	value, err := repo.Client.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return Order{}, ErrNotExist
	} else if err != nil {
		return Order{}, fmt.Errorf("failed to find order: %w", err)
	}

	var order Order
	err = json.Unmarshal([]byte(value), &order)

	if err != nil {
		return Order{}, fmt.Errorf("failed to decode order json: %w", err)
	}

	return order, nil
}

func (repo *RedisRepo) DeleteByID(ctx context.Context, id int64) error {
	key := orderIdKey(id)

	txPipe := repo.Client.TxPipeline()

	err := txPipe.Del(ctx, key).Err()

	if errors.Is(err, redis.Nil) {
		txPipe.Discard()
		return ErrNotExist
	} else if err != nil {
		txPipe.Discard()
		return fmt.Errorf("failed to delete order: %w", err)
	}

	if err := txPipe.SRem(ctx, "orders", key).Err(); err != nil {
		txPipe.Discard()
		return fmt.Errorf("failed to remove from orders set %w", err)
	}

	if _, err := txPipe.Exec(ctx); err != nil {
		txPipe.Discard()
		return fmt.Errorf("failed to exec delete: %w", err)
	}

	return nil
}

func (repo *RedisRepo) Update(ctx context.Context, order Order) error {
	data, err := json.Marshal(order)

	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}

	key := orderIdKey(order.OrderID)

	// Update only existing records.
	err = repo.Client.SetXX(ctx, key, string(data), 0).Err()

	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

type FindResult struct {
	Orders []Order
	Cursor uint64
}

func (repo *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := repo.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get order ids: %w", err)
	}

	// Early return when no keys have been found.
	if len(keys) == 0 {
		return FindResult{Orders: []Order{}, Cursor: cursor}, nil
	}

	unmarshedOrders, err := repo.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get order ids: %w", err)
	}

	orders := make([]Order, len(unmarshedOrders))

	for i, uOrder := range unmarshedOrders {
		uOrder := uOrder.(string)
		var order Order

		err := json.Unmarshal([]byte(uOrder), &order)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode order json: %w", err)
		}

		orders[i] = order
	}

	return FindResult{Orders: orders, Cursor: cursor}, nil
}
