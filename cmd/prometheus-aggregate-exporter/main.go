package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	listenAddress = kingpin.Flag(
		"web.listen-address",
		"Address on which to expose metrics and web interface.",
	).Default(":9100").String()
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	configPath = kingpin.Flag(
		"app.config-path",
		"Path to config YAML file.",
	).Default("config.yml").String()
	sslFlag = kingpin.Flag(
		"web.ssl-transport",
		"Enable SSL HTTP Transport.",
	).Short('s').Bool()
	certFilePath = kingpin.Flag(
		"web.tls-cert-path",
		"Path to TLS cert file.",
	).Default("cert.pem").String()
	keyFilePath = kingpin.Flag(
		"web.tls-key-path",
		"Path to TLS key file.",
	).Default("key.pem").String()
	verboseFlag = kingpin.Flag(
		"app.verbose-log",
		"Enable verbose logs.",
	).Short('v').Bool()
	appendLabel = kingpin.Flag(
		"app.append-label",
		"Append a label, to each metrics family aggregated, with the target/source hostname and port of the original exporter scraped.",
	).Short('l').Bool()
	labelName = kingpin.Flag(
		"app.append-label-name",
		"Label name to use if a target name label is appended to metrics.",
	).Default("ae_source").String()
	targetScrapeTimeout = kingpin.Flag(
		"target.scrape-timeout",
		"Timeout waiting for a target exporter to reply when being scraped.",
	).Default("1000ms").Duration()
)

type Context struct {
	Targets         []string
	authSecretToken string
	targetScrapeTimeout time.Duration
}

type ConfigFileInstance struct {
	Targets []string
}

type TargetScrapeResult struct {
	URL          string
	SecondsTaken float64
	MetricFamily map[string]*io_prometheus_client.MetricFamily
	Error        error
}

// HTTP Client structure to hold target exporter to scrape
// It will also contain a reference to the http.Client used to make request to target exporter
type TargetExporterHttpClient struct {
	httpClient *http.Client
}

// The ParseYAML pointerMethod invokes the (pointer)receiver c *ConfigFileInstance
func (c *ConfigFileInstance) ParseYAML(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}
	return nil
}

func (c *Context) handleIndex(httpResponseWritter http.ResponseWriter, r *http.Request) {
	httpResponseWritter.Write([]byte(`<html>
             <head><title>Prometheus Aggregate Exporter</title></head>
             <body>
             <h1>Prometheus Aggregate Exporter</h1>
             <p><a href='/metrics'>Metrics</a></p>
             <h2>Build</h2>
             <pre>` + version.Info() + ` </pre>
             </body>
             </html>`))
}

func (c *Context) handleMetrics(httpResponseWritter http.ResponseWriter, r *http.Request) {
	aggregator := &TargetExporterHttpClient{httpClient: &http.Client{Timeout: time.Duration(c.targetScrapeTimeout) * time.Millisecond}}
	defer r.Body.Close()
	err := r.ParseForm()
	if err != nil {
		http.Error(httpResponseWritter, "Bad Request", http.StatusBadRequest)
		return
	}
	aggregator.Aggregate(c.Targets, httpResponseWritter)
}

func getConfigFromYml() ConfigFileInstance {
	var config ConfigFileInstance
	log.Infof("Opening config file at path %s.", *configPath)
	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatalf("Failed to open config file at path %s due to error: %s", *configPath, err.Error())
	}
	// Close file before exiting main()
	defer configFile.Close()

	// Read contents of YAML config file
	log.Infof("Reading config file %s content.s", *configPath)
	configData, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file at path %s due to error: %s", *configPath, err.Error())
	}
	log.Infof("Parsing config file %s contents as YAML.", *configPath)
	if err := config.ParseYAML(configData); err != nil {
		log.Fatal(err)
	}
	log.Infof("Closing config file %s.", *configPath)
	log.Infof("Parsed config: % +v", config)
	return config

}
// The Aggregate pointerMethod invokes the (pointer)receiver f *TargetExporterHttpClient
func (f *TargetExporterHttpClient) Aggregate(targets []string, output io.Writer) {

	// Create bi-directional channel of target scrape results.
	targetScrapeResultChannel := make(chan *TargetScrapeResult, 100)
	// Call scrapeTarget as go routine which will execute concurrently with main
	for _, target := range targets {
		go f.scrapeTarget(target, targetScrapeResultChannel)
	}
	// I should make this a worker function
	// Anonymous function that's invoked when a result is set in targetScrapeResultChannel
	func(lenTargets int, resultChanTargetScrape chan *TargetScrapeResult) {

		countResults := 0

		allMetricFamilies := make(map[string]*io_prometheus_client.MetricFamily)
		// Loop thru all channels infinitely.
		for {
			if lenTargets == countResults {
				break
			}
			// Setup select to wait for return from go routines scraping targets.
			select {
			case targetScrapeResult := <-resultChanTargetScrape:
				countResults++
				// Log target scrape failure and move along with next scrape target in targets.
				if targetScrapeResult.Error != nil {
					log.Infof("Target scrape error: %s", targetScrapeResult.Error.Error())
					continue
				}
				// Iterate thru every metric family and append label.
				for metricFamilyName, metricFamily := range targetScrapeResult.MetricFamily {
					if *appendLabel {
						for _, metric := range metricFamily.Metric {
							metric.Label = append(metric.Label, &io_prometheus_client.LabelPair{Name: labelName, Value: &targetScrapeResult.URL})
						}
					}
					if existingMf, ok := allMetricFamilies[metricFamilyName]; ok {
						for _, metric := range metricFamily.Metric {
							existingMf.Metric = append(existingMf.Metric, metric)
						}
					} else {
						allMetricFamilies[*metricFamily.Name] = metricFamily
					}
				}
				if *verboseFlag {
					log.Infof("Success: Target %s scraped in %.3f secs.", targetScrapeResult.URL, targetScrapeResult.SecondsTaken)
				}
			}
		}
		//Instantiate an encoder to encode each line of the result metrics scraped based on the prometheus format
		// defined in expfmt.FmtText .
		encoder := expfmt.NewEncoder(output, expfmt.FmtText)
		for _, metricFamily := range allMetricFamilies {
			encoder.Encode(metricFamily)
		}

	}(len(targets), targetScrapeResultChannel)
}

//The scrapeTarget pointerMethod invokes the (pointer)receiver f *TargetExporterHttpClient
// pushes the raw body response into the channel resultChan
// expects:
// 	- targetURL string of exporter to scrape
// 	- resultChan go channel to update with TargetScrapeResult
func (f *TargetExporterHttpClient) scrapeTarget(targetURL string, resultChan chan *TargetScrapeResult) {

	startTime := time.Now()
	httpClientResponse, httpClientError := f.httpClient.Get(targetURL)
	// Set values into TargetScrapeResult
	result := &TargetScrapeResult{URL: targetURL, SecondsTaken: time.Since(startTime).Seconds(), Error: nil}
	if httpClientResponse != nil {
		result.MetricFamily, httpClientError = parseMetricFamilies(httpClientResponse.Body)
		if httpClientError != nil {
			result.Error = fmt.Errorf("failed to add labels to targetURL %s metrics: %s", targetURL, httpClientError.Error())
			resultChan <- result
			return
		}
	}
	if httpClientError != nil {
		result.Error = fmt.Errorf("failed to fetch URL %s due to error: %s", targetURL, httpClientError.Error())
	}
	resultChan <- result
}

// The parseMetricFamilies function parses metric families from target response body data
// expects:
//  - raw metric family data
// returns:
//  - a map of metric families on successful parse, otherwise nil
//  - an error on parse failure
func parseMetricFamilies(sourceData io.Reader) (map[string]*io_prometheus_client.MetricFamily, error) {
	// Instatiate prometheus parser
	parser := expfmt.TextParser{}
	metricFamiles, err := parser.TextToMetricFamilies(sourceData)
	if err != nil {
		return nil, err
	}
	return metricFamiles, nil
}

func httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	// Run the HTTP request in a goroutine and pass the response to f.
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}
	c := make(chan error, 1)
	go func() { c <- f(client.Do(req)) }()
	select {
	case <-ctx.Done():
		tr.CancelRequest(req)
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func main() {

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("prometheus-aggregate-exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	// Get target config from YAML file
	//config := getConfigFromYml()
	config := getConfigFromYml()

	ctx := &Context{Targets: config.Targets, targetScrapeTimeout: *targetScrapeTimeout}
	// Instantiate HTTP client
	mux := http.NewServeMux()

	mux.HandleFunc("/", ctx.handleIndex)
	mux.HandleFunc(*metricsPath, ctx.handleMetrics)
	log.Infoln("Starting prometheus-aggregate-exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())
	log.Infoln("Listening on", *listenAddress)
	if *sslFlag == true {
		if err := http.ListenAndServeTLS(*listenAddress, *certFilePath, *keyFilePath, mux); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := http.ListenAndServe(*listenAddress, mux); err != nil {
			log.Fatal(err)
		}
	}
}
