package router

import (
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/handlers"
	"go-musthave-diploma-tpl/internal/logger"
	"go-musthave-diploma-tpl/internal/session"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthRouter(t *testing.T) {
	routes := []struct {
		url    string
		method string
	}{
		{"/api/user/orders", "POST"},
		{"/api/user/orders", "GET"},
		{"/api/user/balance", "GET"},
		{"/api/user/balance/withdraw", "POST"},
		{"/api/user/withdrawals", "GET"},
	}

	cnf := config.Init()
	sgr := logger.Init()

	ssp := &session.SessionProvider{
		Config: cnf,
	}
	hdp := &handlers.HandlerProvider{
		Repository: nil,
		Config:     cnf,
		Sugar:      sgr,
		Session:    ssp,
	}
	rtr := RouterProvider{
		Repository: nil,
		Config:     cnf,
		Sugar:      sgr,
		Handler:    hdp,
		Session:    ssp,
	}

	for _, r := range routes {
		t.Run(r.url, func(t *testing.T) {
			ts := httptest.NewServer(rtr.Router())
			defer ts.Close()

			request, err := http.NewRequest(r.method, ts.URL+r.url, nil)
			assert.NoError(t, err)

			response, err := ts.Client().Do(request)
			require.NoError(t, err)
			defer func() {
				if err := response.Body.Close(); err != nil {
					log.Printf("failed to close body: %s", err.Error())
				}
			}()

			assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
		})
	}
}
