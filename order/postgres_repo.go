package order

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/jackc/pgx/v5"
)

type PostgresRepo struct {
	Client *pgx.Conn
}

func (p *PostgresRepo) Insert(ctx context.Context, order Order) error {
	// TODO Convert to transaction in order to add both order and line_items at once.

	orderIdBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(orderIdBytes, order.OrderID)

	args := pgx.NamedArgs{
		"orderId":    orderIdBytes,
		"customerId": order.CustomerID,
		"createdAt":  order.CreatedAt,
	}
	_, err := p.Client.Exec(ctx,
		"INSERT INTO order_store (order_id, customer_id, created_at)"+
			"VALUES (@orderId, @customerId, @createdAt)", args)

	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	for _, item := range order.LineItems {
		_, err := p.Client.Exec(ctx,
			"INSERT INTO line_item (item_id, quantity, price, order_id)"+
				"VALUES ($1, $2, $3, $4)",
			item.ItemID, item.Quantity, item.Price, orderIdBytes)

		if err != nil {
			return fmt.Errorf("failed to insert line item: %w", err)
		}
	}

	return nil
}

func (p *PostgresRepo) FindByID(ctx context.Context, id uint64) (Order, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresRepo) DeleteByID(ctx context.Context, id uint64) error {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresRepo) Update(ctx context.Context, order Order) error {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	//TODO implement me
	panic("implement me")
}
