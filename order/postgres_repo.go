package order

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"time"
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
	args := pgx.NamedArgs{
		"orderId": id,
	}

	rows, err := p.Client.Query(ctx, "SELECT item_id, quantity, price FROM line_item WHERE order_id = @orderId", args)
	if err != nil {
		return Order{}, fmt.Errorf("failed to query line item: %w", err)
	}
	defer func(pgx.Rows) {
		rows.Close()
	}(rows)

	var items []LineItem
	for rows.Next() {
		var itemID uuid.UUID
		var quantity uint
		var price uint
		err = rows.Scan(&itemID, &quantity, &price)
		if err != nil {
			return Order{}, fmt.Errorf("error scanning line_item row: %w", err)
		}

		items = append(items, LineItem{itemID, quantity, price})
	}

	rows.Close()
	if rows.Err() != nil {
		return Order{}, fmt.Errorf("error when closing line_item rows %w", rows.Err())
	}

	row := p.Client.QueryRow(ctx, "SELECT order_id, customer_id, created_at, shipped_at, completed_at "+
		"FROM order_store WHERE order_id = @orderId", args)

	var orderID int64
	var customerID uuid.UUID
	var createdAt *time.Time
	var shippedAt *time.Time
	var completedAt *time.Time
	err = row.Scan(&orderID, &customerID, &createdAt, &shippedAt, &completedAt)
	if err != nil {
		return Order{}, fmt.Errorf("error scanning order row: %w", err)
	}

	utc := time.Time.UTC(*createdAt)
	createdAt = &utc
	if shippedAt != nil {
		utc = time.Time.UTC(*shippedAt)
		shippedAt = &utc
	}
	if completedAt != nil {
		utc = time.Time.UTC(*completedAt)
		completedAt = &utc
	}

	return Order{orderID, customerID, items, createdAt, shippedAt, completedAt}, nil
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
	args := pgx.NamedArgs{
		"orderId":     order.OrderID,
		"createdAt":   order.CreatedAt,
		"shippedAt":   order.ShippedAt,
		"completedAt": order.CompletedAt,
	}
	_, err := p.Client.Exec(ctx,
		"UPDATE order_store SET "+
			"created_at = @createdAt, shipped_at = @shippedAt, completed_at = @completedAt WHERE order_id = @orderId",
		args)

	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

func (p *PostgresRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	//TODO implement me
	panic("implement me")
}
