package dto

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestParsePage(t *testing.T) {
	cases := []struct {
		query    string
		page     int
		pageSize int
	}{
		{"", 1, DefaultPageSize},
		{"page=0", 1, DefaultPageSize},
		{"page=-3&pageSize=abc", 1, DefaultPageSize},
		{"pageSize=0", 1, DefaultPageSize},
		{"page=3&pageSize=50", 3, 50},
		{"pageSize=100000", 1, MaxPageSize},
	}

	var gotPage, gotSize int
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		gotPage, gotSize = ParsePage(c)
		return c.SendStatus(fiber.StatusOK)
	})

	for _, tc := range cases {
		req, err := http.NewRequest(http.MethodGet, "/?"+tc.query, nil)
		if err != nil {
			t.Fatalf("%q: new request: %v", tc.query, err)
		}
		if _, err := app.Test(req); err != nil {
			t.Fatalf("%q: request: %v", tc.query, err)
		}
		if gotPage != tc.page || gotSize != tc.pageSize {
			t.Fatalf("%q: got page=%d pageSize=%d, want page=%d pageSize=%d", tc.query, gotPage, gotSize, tc.page, tc.pageSize)
		}
	}
}
