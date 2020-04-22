package exchange

import (
	"encoding/json"
	"fmt"
	"makerdao/gofer/model"
	"makerdao/gofer/query"
	"strconv"
	"strings"
)

// Bitstamp URL
const bitstampURL = "https://www.bitstamp.net/api/v2/ticker/%s"

type bitstampResponse struct {
	Ask       string `json:"ask"`
	Volume    string `json:"volume"`
	Price     string `json:"last"`
	Bid       string `json:"bid"`
	Timestamp int64  `json:"timestamp"`
}

// Bitstamp exchange handler
type Bitstamp struct{}

// Call implementation
func (b *Bitstamp) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	pair := strings.ToUpper(pp.Pair.Base + pp.Pair.Quote)
	req := &query.HTTPRequest{
		URL: fmt.Sprintf(bitstampURL, pair),
	}

	// make query
	res := pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp bitstampResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to pargse bitstamp response: %s", err)
	}

	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from bitstamp exchange %s", res.Body)
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from bitstamp exchange %s", res.Body)
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(resp.Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from bitstamp exchange %s", res.Body)
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from bitstamp exchange %s", res.Body)
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     model.PriceFromFloat(price),
		Volume:    model.PriceFromFloat(volume),
		Ask:       model.PriceFromFloat(ask),
		Bid:       model.PriceFromFloat(bid),
		Timestamp: resp.Timestamp,
	}, nil
}
