package orderservice

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetOrder(t *testing.T) {
	id := "123"
	some := "test"
	status := "ok"
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/order/%s?some=%s", id, some), nil)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/order/{ID}", GetOrder).Methods(http.MethodGet)
	router.ServeHTTP(w, req)

	response := w.Result()
	if response.StatusCode != http.StatusOK {
		t.Errorf("Status code is wrong. Have: %d, want: %d.", response.StatusCode, http.StatusOK)
	}

	var resp OrderResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}
	if resp.Status != status {
		t.Errorf("Expected status 'ok', got %q", resp.Status)
	}
	if resp.ID != id {
		t.Errorf("Expected ID '123', got %q", resp.ID)
	}
	if resp.Some != some {
		t.Errorf("Expected Some 'test', got %q", resp.Some)
	}
}
