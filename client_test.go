package tc4400exporter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.FileServer(http.Dir("./testdata")))
	defer server.Close()

	client := Client{
		HTTPClient: http.DefaultClient,
		RootURL:    server.URL,
		Username:   "test",
		Password:   "test",
	}

	info, err := client.Info(ctx)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%+v\n", info)

	stats, err := client.Stats(ctx)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%+v\n", stats)
}
