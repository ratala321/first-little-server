package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	Repo Repository
}

type Repository interface {
	Insert(ctx context.Context, order Order) error
	FindByID(ctx context.Context, id int64) (Order, error)
	DeleteByID(ctx context.Context, id int64) error
	Update(ctx context.Context, order Order) error
	FindAll(ctx context.Context, page FindAllPage) (FindResult, error)
}

type FindAllPage struct {
	Size   uint
	Offset uint64
}

type FindResult struct {
	Orders []Order
	Cursor uint64
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID uuid.UUID  `json:"customer_id"`
		LineItems  []LineItem `json:"line_items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	createdOrder := Order{
		OrderID:    rand.Int63(),
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}

	err := h.Repo.Insert(r.Context(), createdOrder)
	if err != nil {
		fmt.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(createdOrder)
	if err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(res)
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	const decimal = 10
	const bitSize = 64
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		fmt.Println("failed to parse cursor:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const size = 50
	res, err := h.Repo.FindAll(r.Context(), FindAllPage{Size: size, Offset: cursor})
	if err != nil {
		fmt.Println("failed to find:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Items []Order `json:"items"`
		Next  uint64  `json:"next,omitempty"`
	}
	response.Items = res.Orders
	response.Next = res.Cursor

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(data)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseInt(idParam, base, bitSize)
	if err != nil {
		fmt.Println("failed to parse id:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	found, err := h.Repo.FindByID(r.Context(), orderID)
	if errors.Is(err, ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(found); err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateByID(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fmt.Println("failed to parse body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idParam := chi.URLParam(r, "id")
	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseInt(idParam, base, bitSize)
	if err != nil {
		fmt.Println("failed to parse id:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	toUpdate, err := h.Repo.FindByID(r.Context(), orderID)
	if errors.Is(err, ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	const completedStatus = "completed"
	const shippedStatus = "shipped"
	now := time.Now().UTC()

	switch body.Status {
	case completedStatus:
		if toUpdate.CompletedAt != nil || toUpdate.ShippedAt == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		toUpdate.CompletedAt = &now
		break
	case shippedStatus:
		if toUpdate.ShippedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		toUpdate.ShippedAt = &now
		break
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.Repo.Update(r.Context(), toUpdate); err != nil {
		fmt.Println("failed to update:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(toUpdate); err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64
	orderID, err := strconv.ParseInt(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.Repo.DeleteByID(r.Context(), orderID)
	if errors.Is(err, ErrNotExist) {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if err != nil {
		fmt.Println("failed to delete by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
