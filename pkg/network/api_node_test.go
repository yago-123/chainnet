package network

import (
	"github.com/julienschmidt/httprouter"
	"github.com/yago-123/chainnet/config"
	"github.com/yago-123/chainnet/tests/mocks/chain/explorer"
	"github.com/yago-123/chainnet/tests/mocks/encoding"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListTransactions(t *testing.T) {
	router := NewHTTPRouter(config.NewConfig(), &encoding.MockEncoding{}, &explorer.MockExplorer{}, nil)

	// Create a test request (method and URL can be adjusted)
	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)

	params := httprouter.Params{httprouter.Param{Key: "address", Value: "test"}}

	w := httptest.NewRecorder()
	router.listTransactions(w, req, params)

}
