package gatherers

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/google/uuid"
	"github.com/johnmillner/robo-macd/internal/transformers"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
	"time"
)

type FetcherConfig struct {
	To         uuid.UUID
	Active     bool
	SimpleData chan transformers.SimpleData
	Client     alpaca.Client
	Symbols    []string
	Period     time.Duration
	Limit      int
}

func (c FetcherConfig) GetTo() uuid.UUID {
	return c.To
}

func (c FetcherConfig) IsActive() bool {
	return c.Active
}

func fetchData(config FetcherConfig) {

	values, err := config.Client.ListBars(config.Symbols, alpaca.ListBarParams{
		Timeframe: durationToTimeframe(config.Period),
		Limit:     &config.Limit,
	})

	if err != nil {
		log.Printf("could not fetch bars from alpaca due to %s", err)
		return
	}

	for symbol, bars := range values {
		bars := bars
		symbol := symbol
		go func() {
			for _, bar := range bars {
				config.SimpleData <- transformers.SimpleData{
					Product: symbol,
					Time:    bar.GetTime(),
					Price:   bar.Close,
				}
			}
		}()
	}

}

func durationToTimeframe(dur time.Duration) string {
	switch dur {
	case time.Minute:
		return "1Min"
	case time.Minute * 5:
		return "5Min"
	case time.Minute * 15:
		return "15Min"
	case time.Hour * 24:
		return "1D"
	default:
		log.Fatalf("cannot translate duration given to alpaca timeframe, given: %f (in seconds) - only acceptable durations are 1min, 5min, 15min, 1day", dur.Seconds())
		return dur.String()
	}
}

func StartFetching(configManager utils.ConfigManager) {
	for config := configManager.Check().(FetcherConfig); config.Active; time.Sleep(config.Period) {
		fetchData(config)
	}
}
