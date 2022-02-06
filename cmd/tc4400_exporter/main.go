// Command arris_exporter implements a Prometheus exporter for Arris cable
// modem devices.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	tc4400exporter "github.com/blainsmith/tc4400_exporter"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Listen struct {
		Addr        string `yaml:"address"`
		MetricsPath string `yaml:"metricspath"`
	} `yaml:"listen"`
	Modems []struct {
		Addr     string        `yaml:"address"`
		Username string        `yaml:"username"`
		Password string        `yaml:"password"`
		Timeout  time.Duration `yaml:"timeout"`
	} `yaml:"modems"`
}

func main() {
	var configFile = flag.String("config.file", "", "Relative path to config file yaml")
	flag.Parse()

	var config Config
	source, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("failed to read config file %q: %v", *configFile, err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		log.Fatalf("failed to read YAML from config file %q: %v", *configFile, err)
	}

	// dial is the function used to connect to an Arris device on each
	// metrics scrape request.
	dial := func(addr string) (*tc4400exporter.Client, error) {
		for _, modem := range config.Modems {
			if addr == modem.Addr {
				return &tc4400exporter.Client{
					HTTPClient: &http.Client{
						Timeout: modem.Timeout,
					},
					RootURL:  modem.Addr,
					Username: modem.Username,
					Password: modem.Password,
				}, nil
			}
		}

		return nil, fmt.Errorf("%s not configured", addr)
	}

	h := tc4400exporter.NewHandler(dial)

	mux := http.NewServeMux()
	mux.Handle(config.Listen.MetricsPath, h)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, config.Listen.MetricsPath, http.StatusMovedPermanently)
	})

	log.Printf("starting TC4400 exporter on %q", config.Listen.Addr)

	if err := http.ListenAndServe(config.Listen.Addr, mux); err != nil {
		log.Fatalf("cannot start TC4400 exporter: %v", err)
	}
}
