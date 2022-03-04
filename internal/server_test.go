package internal

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	ctx := context.Background()
	sigs := make(chan os.Signal, 1)

	addr := StartServer(ctx, sigs, "")

	time.Sleep(500 * time.Millisecond)

	client := http.Client{Timeout: 5 * time.Second}

	res, err := client.Get("http://" + addr + "/metrics")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	if res != nil {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	}
	if res != nil {
		res.Body.Close()
	}
	sigs <- os.Interrupt
}

func BenchmarkStartServer(b *testing.B) {
	client := http.Client{Timeout: 5 * time.Second}
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			res, err := client.Get("http://127.0.0.1:2114/metrics")
			if err != nil {
				b.Error(err)
			}
			if res != nil && res.StatusCode != http.StatusOK {
				b.Error("status code: " + res.Status)
			}
			if res != nil {
				res.Body.Close()
			}
		}
	})
}
