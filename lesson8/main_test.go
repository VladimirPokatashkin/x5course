package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// До реализации health этот тест должен быть красным.
func TestHealth(t *testing.T) {
	h := newRouter(NewStore())
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type=%q", rec.Header().Get("Content-Type"))
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" {
		t.Fatalf("status in body=%q", body.Status)
	}
}

func TestCreateOrder(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"valid", `{customer: "iluxa"}`},
		{"empty body", ""},
		{"whitespace body", " \n\t"},
		{"invalid JSON", `{"customer":`},
		{"invalid customer type", `{"customer":123}`},
		{"missing customer", `{}`},
		{"empty name", `{"customer":""}`},
		{"whitespace name", `{"customer":" \t\n "}`},
		{"null", `null`},
		{"null customer", `{"customer":null}`},
		{"array", `[]`},
		{"second JSON", `{"customer":"timur"} {}`},
		{"trailing garbage", `{"customer":"yadykin"} aboba`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore()
			existing := store.Create("vova")
			h := newRouter(store)
			req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status=%d, want 400", rec.Code)
			}
			if rec.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("Content-Type=%q", rec.Header().Get("Content-Type"))
			}
			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error == "" {
				t.Fatal("empty error message")
			}
			if len(store.orders) != 1 || store.nextID != 2 || store.orders[existing.ID] != existing {
				t.Fatal("invalid request changed the store")
			}
		})
	}
}
