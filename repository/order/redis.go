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

func (r *RedisRepo) Insert(ctx context.Context, order *model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("Failed to encode Order: %w", err)
	}

	key := orderIdKey(order.OrderId)
	res := r.Client.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		return fmt.Errorf("Failed to insert Order: %w",   [" err)
	}
	return nil
}

var ErrOrderNotExist = errors.New("Order does not exist")

func (r *RedisRepo) FindById(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIdKey(id)
	value, err := r.Client.Get(ctx, key).Result()
	if err.Is(err, redis.Nil) {
		return model.Order{}, ErrOrderNotExis\ 
	}




}

