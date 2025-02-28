package order_test

import (
	"context"
	"encoding/json"
	"first-little-server/order"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const postgresAddressFilePath = "./testdata/postgres_test_address.txt"

func TestGetByID(t *testing.T) {
	ctx := context.Background()

	res := sendGetByIDRequest(t, ctx)
	defer res.Body.Close()

	actual := getOrderJsonFromResponse(t, res)

	const expected = `{"order_id":5195209675646023232,"customer_id":"0af92773-2324-4788-8a07-eaf0a93b8b83","line_items":[{"item_id":"a0fbd434-fca0-40d5-92dc-6ab5a5bac947","quantity":5,"price":1999},{"item_id":"3d5d6e4e-d404-4e9a-82ff-f8ae9af479e0","quantity":3,"price":2000},{"item_id":"b7764a22-143a-4fe7-bd1d-cc1d0e701d41","quantity":2,"price":3}],"created_at":"2025-02-25T22:02:58.087077Z","shipped_at":"2025-02-25T22:02:58.087077Z","completed_at":"2025-02-25T22:02:58.087077Z"}`

	if string(actual) != expected {
		t.Fatal("expected", expected, "got", actual)
	}
}

func connectToDatabase(t *testing.T, ctx context.Context) *pgx.Conn {
	file, err := os.ReadFile(postgresAddressFilePath)
	if err != nil {
		t.Fatal("error reading file containing postgres test address:", err)
	}

	var address struct {
		Test string `json:"test"`
	}

	_ = json.Unmarshal(file, &address)

	postgres, err := pgx.Connect(ctx, address.Test)
	if err != nil {
		t.Fatal("error connecting to postgres:", err)
	}
	return postgres
}

func sendGetByIDRequest(t *testing.T, ctx context.Context) *http.Response {
	postgres := connectToDatabase(t, ctx)

	handler := order.Handler{
		Repo: &order.PostgresRepo{
			Client: postgres,
		},
	}

	router := chi.NewRouter()
	router.Route("/orders", func(r chi.Router) {
		r.Get("/{id}", handler.GetByID)
	})

	req := httptest.NewRequest(http.MethodGet, "/orders/5195209675646023232", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	res := recorder.Result()
	return res
}

func getOrderJsonFromResponse(t *testing.T, res *http.Response) []byte {
	var bodyOrder order.Order
	err := json.NewDecoder(res.Body).Decode(&bodyOrder)
	if err != nil {
		t.Fatal("error decoding body:", err)
	}

	actual, err := json.Marshal(res.Body)
	if err != nil {
		t.Fatal("error marshalling back to json:", err)
	}
	return actual
}
