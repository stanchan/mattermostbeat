package beater

import (
	"fmt"
	"time"
	"io/ioutil"
	"net/http"
	"bufio"
	"strings"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/stanchan/mattermostbeat/config"
)

type Mattermostbeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Mattermostbeat{
		done:   make(chan struct{}),
		config: config,
	}
	return bt, nil
}

func (bt *Mattermostbeat) Run(b *beat.Beat) error {
	logp.Info("mattermostbeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	for {
		bt.sendGetMetrics(bt.config.Scheme, bt.config.Host, bt.config.Port, b.Name) // call sendGetMetrics
		logp.Info("Event sent")
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}
	}
}

func (bt *Mattermostbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

func (bt *Mattermostbeat) sendGetMetrics(remoteScheme string, remoteHost string, remotePort int, beatname string) {
	// Get Metrics (GET http://localhost:8067/metrics)

	// Create client
	client := &http.Client{}

	urlString := fmt.Sprintf("%s://%s:%d/metrics", remoteScheme, remoteHost, remotePort)

	// Create request
	req, err := http.NewRequest("GET", urlString, nil)

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		logp.Critical("Failure: %s", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Display HTTP Raw Results
	//fmt.Println("response Status : ", resp.Status)
	//fmt.Println("response Headers : ", resp.Header)
	//fmt.Println("response Body : ", string(respBody))

	// Create event
	event := common.MapStr{
		"@timestamp": common.Time(time.Now()),
		"type":       beatname,
	}

	// Append space delimited payload while skipping comments
	scanner := bufio.NewScanner(strings.NewReader(string(respBody)))
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "#"); equal < 0 {
			s := strings.Split(line, " ")
			key, value := s[0], s[1]
			event[key] = value
		}
	}

	// Display raw hash
	//for key, value := range event {
	//	fmt.Printf("key[%s] value[%s]\n", key, value)
	//}

	// Send event
	bt.client.PublishEvent(event)
}
