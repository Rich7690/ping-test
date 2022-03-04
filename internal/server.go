package internal

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

type Endpoint struct {
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
}
type Config struct {
	LogLevel  string     `yaml:"logLevel,omitempty" json:"logLevel,omitempty"`
	Endpoints []Endpoint `yaml:"endpoints,omitempty" json:"endpoints,omitempty"`
}

type IPConfig struct {
	Ip string `json:"ip,omitempty"`
}

var ip = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "stats_ip",
	Help: "ip address",
}, []string{"ip"})

var queryTime = promauto.NewHistogram(prometheus.HistogramOpts{
	Name: "stats_query_duration",
	Help: "stats query duration",
})
var client *http.Client

func StartServer(ctx context.Context, sigs chan os.Signal, addr string) string {

	rclient := retryablehttp.NewClient()
	rclient.RetryMax = 3
	rclient.RetryWaitMax = 500 * time.Millisecond
	rclient.RetryWaitMin = 100 * time.Millisecond
	rclient.HTTPClient.Timeout = 2 * time.Second
	rclient.Logger = nil
	client = rclient.StandardClient()

	ro := mux.NewRouter()
	ro.HandleFunc("/metrics", GetMetricsHandler())
	ro.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) { rw.WriteHeader(http.StatusOK) })
	ro.HandleFunc("/debug/pprof/", pprof.Index)
	ro.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	ro.HandleFunc("/debug/pprof/profile", pprof.Profile)
	ro.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	ro.HandleFunc("/debug/pprof/trace", pprof.Trace)
	ro.HandleFunc("/debug/pprof/{name}", func(rw http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		pprof.Handler(params["name"]).ServeHTTP(rw, r)
	})

	server := http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      ro,
	}

	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open listener")
	}

	go func() {
		err := server.Serve(ln)
		if err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Error listening server")
		}
	}()
	log.Info().Str("addr", ln.Addr().String()).Msg("Listening")
	return ln.Addr().String()
}

func GetMetricsHandler() func(rw http.ResponseWriter, r *http.Request) {
	handler := promhttp.Handler()
	return func(rw http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "https://checkip.amazonaws.com", nil)
		if err != nil {
			log.Err(err).Msg("Error creating request")
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		start := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			log.Err(err).Msg("Error making request")
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		dur := time.Since(start)
		defer resp.Body.Close()
		if resp.ContentLength < 0 {
			log.Error().Msg("content length was 0")
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		buf := make([]byte, resp.ContentLength-1)

		_, err = resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			log.Err(err).Msg("Error reading body")
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		ip.Reset()
		ip.WithLabelValues(strings.TrimSpace(string(buf))).Set(1)
		queryTime.Observe(dur.Seconds())

		handler.ServeHTTP(rw, r)
	}
}
