package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/anand-aman/orders-api/model"
	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func orderIdKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("Failed to encode Order: %w", err)
	}

	key := orderIdKey(order.OrderId)

	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("Failed to insert Order: %w", err)
	}

	//Add order Id in a set
	resSAdd := txn.SAdd(ctx, "orders", key)
	if resSAdd.Err() != nil {
		txn.Discard()
		return fmt.Errorf("Failed to add to the Orders set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("Failed to execute insert transaction: %w", err)
	}

	return nil
}

var ErrOrderNotExist = errors.New("Order does not exist")

func (r *RedisRepo) FindById(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIdKey(id)
	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrOrderNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("Error occured during Order fetch: %w", err)
	}

	var order model.Order
	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("Failed to decode order json: %w", err)
	}

	return order, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := orderIdKey(id)

	txn := r.Client.TxPipeline()
	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrOrderNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("get order: %w", err)
	}

	//Remove OrderId from the Order Id Set
	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("Failed to remove from Orders set: %w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}

	key := orderIdKey(order.OrderId)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrOrderNotExist
	} else if err != nil {
		return fmt.Errorf("Failed to update Order: %w", err)
	}

	return nil
}

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("Failed to get Order Ids: %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("Failed to get Orders: %w", err)
	}

	orders := make([]model.Order, len(xs))
	for i, x := range xs {
		x := x.(string)
		var order model.Order

		err := json.Unmarshal([]byte(x), &order)
		if err != nil {
			return FindResult{}, fmt.Errorf("Failed to decode Order Json: %w", err)
		}
		orders[i] = order
	}

	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil

}
