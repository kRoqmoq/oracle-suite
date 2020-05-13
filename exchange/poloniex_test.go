//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package exchange

import (
	"fmt"
	"makerdao/gofer/model"
	"makerdao/gofer/query"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type PoloniexSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange Handler
}

// Setup exchange
func (suite *PoloniexSuite) SetupSuite() {
	suite.exchange = &Poloniex{}
}

func (suite *PoloniexSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *PoloniexSuite) TestFailOnWrongInput() {
	// no pool
	_, err := suite.exchange.Call(nil, nil)
	suite.Equal(errNoPoolPassed, err)

	// empty pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), nil)
	suite.Error(err)

	// wrong pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), &model.PotentialPricePoint{})
	suite.Error(err)

	pp := newPotentialPricePoint("poloniex", "BTC", "ETH")
	// nil as response
	_, err = suite.exchange.Call(newMockWorkerPool(nil), pp)
	suite.Equal(errEmptyExchangeResponse, err)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Equal(ourErr, err)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte("{}"),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(`{"BBB_EEE":{}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"ETH_BTC":{"last":"abc"}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"ETH_BTC":{"last":"1","lowestAsk":"abc"}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"ETH_BTC":{"last":"1","lowestAsk":"1","baseVolume":"abc"}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"ETH_BTC":{"last":"1","lowestAsk":"1","baseVolume":"1","highestBid":"abc"}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"BBB_ETT":{"last":"1","lowestAsk":"1","baseVolume":"1","highestBid":"1"}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)
}

func (suite *PoloniexSuite) TestSuccessResponse() {
	pp := newPotentialPricePoint("poloniex", "BTC", "ETH")
	resp := &query.HTTPResponse{
		Body: []byte(`{"ETH_BTC":{"last":"1","lowestAsk":"2","baseVolume":"3","highestBid":"4"}}`),
	}
	point, err := suite.exchange.Call(newMockWorkerPool(resp), pp)

	suite.NoError(err)
	suite.Equal(pp.Exchange, point.Exchange)
	suite.Equal(pp.Pair, point.Pair)
	suite.Equal(model.PriceFromFloat(1.0), point.Price)
	suite.Equal(model.PriceFromFloat(2.0), point.Ask)
	suite.Equal(model.PriceFromFloat(3.0), point.Volume)
	suite.Equal(model.PriceFromFloat(4.0), point.Bid)
	suite.Greater(point.Timestamp, int64(2))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestPoloniexSuite(t *testing.T) {
	suite.Run(t, new(PoloniexSuite))
}
