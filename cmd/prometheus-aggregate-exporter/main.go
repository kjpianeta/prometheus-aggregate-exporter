package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"gopkg.in/yaml.v2"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	configPathFlag = flag.String("config", "config.yml", "Path to config YAML file.")
	verboseFlag    = flag.Bool("verbose", false, "Log more information")
	versionFlag    = flag.Bool("version", false, "Show version and exit")
	appendLabel    = flag.Bool("label", true, "Add a label to metrics to show their origin target")
	labelName      = flag.String("label.name", "ae_source", "Label name to use if a target name label is appended to metrics")
)

type instanceConfig struct {
	Server struct {
		Bind              string
		HttpTransport     string
		Verbose           bool
		AppendSourceLabel bool
		SourceLabelName   string
	}
	Timeout int
	Targets []string
}

type Result struct {
	URL          string
	SecondsTaken float64
	MetricFamily map[string]*io_prometheus_client.MetricFamily
	Error        error
}

type Aggregator struct {
	HTTP *http.Client
}

func (c *instanceConfig) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}
	if c.Server.Bind == "" {
		return errors.New("Prometheus aggregator config: invalid bind.")
	}
	return nil
}

func main() {

	flag.Parse()

	if *versionFlag {
		fmt.Printf("prometheus-aggregate-exporter: %v, commit %v, built at %v", version, commit, date)
		os.Exit(0)
	}
	configFile, err := os.Open(*configPathFlag)
	if err != nil {
		log.Fatalf("Failed to open config file at path %s due to error: %s", *configPathFlag, err.Error())
	}
	defer configFile.Close()

	configData, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file at path %s due to error: %s", *configPathFlag, err.Error())
	}
	var config instanceConfig
	if err := config.Parse(configData); err != nil {
		log.Fatal(err)
	}

	aggregator := &Aggregator{HTTP: &http.Client{Timeout: time.Duration(config.Timeout) * time.Millisecond}}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		err := r.ParseForm()
		if err != nil {
			http.Error(rw, "Bad Request", http.StatusBadRequest)
			return
		}
		if t := r.Form.Get("t"); t != "" {
			targetKey, err := strconv.Atoi(t)
			if err != nil || len(config.Targets)-1 < targetKey {
				http.Error(rw, "Bad Request", http.StatusBadRequest)
				return
			}
			aggregator.Aggregate([]string{config.Targets[targetKey]}, rw)
		} else {
			aggregator.Aggregate(config.Targets, rw)
		}
	})

	log.Printf("Starting server on %s...", config.Server.Bind)
	log.Fatal(http.ListenAndServe(config.Server.Bind, mux))
}

func (f *Aggregator) Aggregate(targets []string, output io.Writer) {

	resultChan := make(chan *Result, 100)

	for _, target := range targets {
		go f.fetch(target, resultChan)
	}

	func(numTargets int, resultChan chan *Result) {

		numResuts := 0

		allFamilies := make(map[string]*io_prometheus_client.MetricFamily)

		for {
			if numTargets == numResuts {
				break
			}
			select {
			case result := <-resultChan:
				numResuts++

				if result.Error != nil {
					log.Printf("Fetch error: %s", result.Error.Error())
					continue
				}

				for mfName, mf := range result.MetricFamily {
					if *appendLabel {
						for _, m := range mf.Metric {
							m.Label = append(m.Label, &io_prometheus_client.LabelPair{Name: labelName, Value: &result.URL})
						}
					}
					if existingMf, ok := allFamilies[mfName]; ok {
						for _, m := range mf.Metric {
							existingMf.Metric = append(existingMf.Metric, m)
						}
					} else {
						allFamilies[*mf.Name] = mf
					}
				}
				if *verboseFlag {
					log.Printf("OK: %s was refreshed in %.3f seconds", result.URL, result.SecondsTaken)
				}
			}
		}

		encoder := expfmt.NewEncoder(output, expfmt.FmtText)
		for _, f := range allFamilies {
			encoder.Encode(f)
		}

	}(len(targets), resultChan)
}

func (f *Aggregator) fetch(target string, resultChan chan *Result) {

	startTime := time.Now()
	res, err := f.HTTP.Get(target)

	result := &Result{URL: target, SecondsTaken: time.Since(startTime).Seconds(), Error: nil}
	if res != nil {
		result.MetricFamily, err = getMetricFamilies(res.Body)
		if err != nil {
			result.Error = fmt.Errorf("failed to add labels to target %s metrics: %s", target, err.Error())
			resultChan <- result
			return
		}
	}
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch URL %s due to error: %s", target, err.Error())
	}
	resultChan <- result
}

func getMetricFamilies(sourceData io.Reader) (map[string]*io_prometheus_client.MetricFamily, error) {
	parser := expfmt.TextParser{}
	metricFamiles, err := parser.TextToMetricFamilies(sourceData)
	if err != nil {
		return nil, err
	}
	return metricFamiles, nil
}
