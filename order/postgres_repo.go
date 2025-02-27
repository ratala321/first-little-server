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

const (
	orderTable = "order_store"

	orderIdRow     = "order_id"
	customerIdRow  = "customer_id"
	createdAtRow   = "created_at"
	shippedAtRow   = "shipped_at"
	completedAtRow = "completed_at"

	lineItemTable = "line_item"

	lineItemIdRow = "item_id"
	quantityRow   = "quantity"
	priceRow      = "price"
)

const insertIntoOrderSQL = "INSERT INTO " + orderTable +
	" (" + orderIdRow + ", " + customerIdRow + ", " + createdAtRow + ")" +
	" VALUES (@orderId, @customerId, @createdAt)"
const insertIntoLineItemSQL = "INSERT INTO " + lineItemTable +
	" (" + lineItemIdRow + ", " + quantityRow + ", " + priceRow + ", " + orderIdRow + ")" +
	"VALUES ($1, $2, $3, $4)"

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
	_, err = tx.Exec(ctx, insertIntoOrderSQL, args)

	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	for _, item := range order.LineItems {
		_, err := tx.Exec(ctx, insertIntoLineItemSQL, item.ItemID, item.Quantity, item.Price, order.OrderID)

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

const selectLineItemSQL = "SELECT " + lineItemIdRow + ", " + quantityRow + ", " + priceRow +
	" FROM " + lineItemTable + " WHERE " + orderIdRow + " = @orderId"
const selectOrderSQL = "SELECT " + orderIdRow + ", " + customerIdRow + ", " + createdAtRow +
	", " + shippedAtRow + ", " + completedAtRow + " " +
	"FROM " + orderTable + " WHERE " + orderIdRow + " = @orderId"

func (p *PostgresRepo) FindByID(ctx context.Context, id int64) (Order, error) {
	args := pgx.NamedArgs{
		"orderId": id,
	}

	rows, err := p.Client.Query(ctx, selectLineItemSQL, args)
	defer func(pgx.Rows) {
		rows.Close()
	}(rows)
	if err != nil {
		return Order{}, fmt.Errorf("failed to query line item: %w", err)
	}

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

	row := p.Client.QueryRow(ctx, selectOrderSQL, args)

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

const deleteLineItemSQL = "DELETE FROM " + lineItemTable + " WHERE " + orderIdRow + " = @orderId"
const deleteOrderSQL = "DELETE FROM " + orderTable + " WHERE " + orderIdRow + " = @orderId"

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

	_, err = tx.Exec(ctx, deleteLineItemSQL, args)
	if err != nil {
		return fmt.Errorf("failed to delete associated line item: %w", err)
	}

	_, err = tx.Exec(ctx, deleteOrderSQL, args)

	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit order transaction %w", err)
	}

	return nil
}

const updateOrderSQL = "UPDATE " + orderTable + " SET " +
	createdAtRow + " = @createdAt, " + shippedAtRow + " = @shippedAt, " +
	completedAtRow + " = @completedAt WHERE " + orderIdRow + " = @orderId"

func (p *PostgresRepo) Update(ctx context.Context, order Order) error {
	args := pgx.NamedArgs{
		"orderId":     order.OrderID,
		"createdAt":   order.CreatedAt,
		"shippedAt":   order.ShippedAt,
		"completedAt": order.CompletedAt,
	}
	_, err := p.Client.Exec(ctx, updateOrderSQL, args)

	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

const findAllSQL = "SELECT os." + orderIdRow + ", " + customerIdRow + ", " + createdAtRow + ", " + shippedAtRow +
	", " + completedAtRow + ", " + lineItemIdRow + ", " + quantityRow + ", " + priceRow + " " +
	"FROM " + orderTable + " AS os JOIN " + lineItemTable + " AS li " +
	"ON os." + orderIdRow + " = li." + orderIdRow + " OFFSET $1 LIMIT $2"

func (p *PostgresRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	rows, err := p.Client.Query(ctx, findAllSQL, page.Offset, page.Size)
	defer func(pgx.Rows) {
		rows.Close()
	}(rows)

	if err != nil {
		return FindResult{}, fmt.Errorf("failed to query line item: %w", err)
	}

	var orders []Order
	const invalidOrderID int64 = -1
	var lastOrderID int64 = -1
	for rows.Next() {
		var (
			orderID     int64
			customerID  uuid.UUID
			createdAt   *time.Time
			shippedAt   *time.Time
			completedAt *time.Time
		)
		var (
			itemID   uuid.UUID
			quantity uint
			price    uint
		)

		err = rows.Scan(&orderID, &customerID, &createdAt, &shippedAt, &completedAt, &itemID, &quantity, &price)

		if err != nil {
			return FindResult{}, fmt.Errorf("error scanning orders join on line_item row: %w", err)
		}

		if orderID == invalidOrderID || orderID != lastOrderID {
			lastOrderID = orderID
			orders = append(orders, Order{orderID, customerID, []LineItem{}, createdAt, shippedAt, completedAt})
		}

		orders[len(orders)-1].LineItems = append(orders[len(orders)-1].LineItems, LineItem{itemID, quantity, price})
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		return FindResult{}, fmt.Errorf("error closing rows: %w", err)
	}

	return FindResult{Orders: orders, Cursor: page.Offset + uint64(page.Size)}, nil
}
