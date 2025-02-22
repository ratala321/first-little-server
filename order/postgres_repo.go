package order

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

type PostgresRepo struct {
	Client *pgx.Conn
}

func (p *PostgresRepo) Insert(ctx context.Context, order Order) error {
	tx, err := p.Client.BeginTx(ctx, pgx.TxOptions{})
	defer func(context.Context, pgx.Tx, error) {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}(ctx, tx, err)

	if err != nil {
		return fmt.Errorf("failed to begin transaction for order: %w", err)
	}

	args := pgx.NamedArgs{
		"orderId":    order.OrderID,
		"customerId": order.CustomerID,
		"createdAt":  order.CreatedAt,
	}
	_, err = tx.Exec(ctx,
		"INSERT INTO order_store (order_id, customer_id, created_at)"+
			"VALUES (@orderId, @customerId, @createdAt)", args)

	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	for _, item := range order.LineItems {
		_, err := tx.Exec(ctx,
			"INSERT INTO line_item (item_id, quantity, price, order_id)"+
				"VALUES ($1, $2, $3, $4)",
			item.ItemID, item.Quantity, item.Price, order.OrderID)

		if err != nil {
			return fmt.Errorf("failed to insert line item: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit order transaction %w", err)
	}

	return nil
}

func (p *PostgresRepo) FindByID(ctx context.Context, id int64) (Order, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresRepo) DeleteByID(ctx context.Context, id int64) error {
	tx, err := p.Client.BeginTx(ctx, pgx.TxOptions{})
	defer func(context.Context, pgx.Tx, error) {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}(ctx, tx, err)

	if err != nil {
		return fmt.Errorf("failed to begin transaction for order: %w", err)
	}

	args := pgx.NamedArgs{
		"orderId": id,
	}

	_, err = tx.Exec(ctx, "DELETE FROM line_item WHERE order_id = @orderId", args)
	if err != nil {
		return fmt.Errorf("failed to delete associated line item: %w", err)
	}

	_, err = tx.Exec(ctx, "DELETE FROM order_store WHERE order_id = @orderId", args)

	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit order transaction %w", err)
	}

	return nil
}

func (p *PostgresRepo) Update(ctx context.Context, order Order) error {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	//TODO implement me
	panic("implement me")
}
