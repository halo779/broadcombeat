package beater

import (
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/halo779/broadcombeat/broadcom"
	"github.com/halo779/broadcombeat/config"
)

type Broadcombeat struct {
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

	bt := &Broadcombeat{
		done:   make(chan struct{}),
		config: config,
	}
	fmt.Println("Beat Created")
	return bt, nil
}

func (bt *Broadcombeat) Run(b *beat.Beat) error {
	logp.Info("broadcombeat is running! Hit CTRL-C to stop it.")
	fmt.Println("broadcombeat is running! Hit CTRL-C to stop it.")
	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		event := common.MapStr{
			"@timestamp": common.Time(time.Now()),
			"type":       b.Name,
		}
		logp.Info("Event sent")
		fmt.Println("Event sent")
		br := broadcom.Process(event)

		fmt.Println(br)
		bt.client.PublishEvent(br)
	}
}

func (bt *Broadcombeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
