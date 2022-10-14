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

package origins

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chronicleprotocol/oracle-suite/pkg/util/query"
)

// Fixer URL
const fixerURL = "https://api.apilayer.com/fixer/latest?base=%s&symbols=$s"

type fixerResponse struct {
	Rates map[string]float64 `json:"rates"`
}

// Fixer exchange handler
type Fixer struct {
	WorkerPool query.WorkerPool
	BaseURL    string
	APIKey     string
}

func (f *Fixer) renameSymbol(symbol string) string {
	return strings.ToUpper(symbol)
}

func (f Fixer) Pool() query.WorkerPool {
	return f.WorkerPool
}

func (f Fixer) PullPrices(pairs []Pair) []FetchResult {
	// Group pairs by asset pair base.
	bases := map[string][]Pair{}
	for _, pair := range pairs {
		base := pair.Base
		bases[base] = append(bases[base], pair)
	}

	var results []FetchResult
	for base, pairs := range bases {
		// Make one request per asset pair base.
		crs, err := f.callByBase(base, pairs)
		if err != nil {
			// If callByBase fails wholesale, create a FetchResult per pair with the same
			// error.
			crs = fetchResultListWithErrors(pairs, err)
		}
		results = append(results, crs...)
	}

	return results
}

func (f *Fixer) getURL(base string, quotes []Pair) string {
	symbols := []string{}
	for _, pair := range quotes {
		symbols = append(symbols, f.renameSymbol(pair.Quote))
	}
	return fmt.Sprintf(fixerURL, strings.Join(symbols, ","), f.renameSymbol(base), f.APIKey)
}

func (f *Fixer) callByBase(base string, pairs []Pair) ([]FetchResult, error) {
	req := &query.HTTPRequest{
		URL: f.getURL(base, pairs),
	}

	// Make query.
	res := f.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// Parse JSON.
	var resp fixerResponse
	err := json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Fixer response: %w", err)
	}
	if resp.Rates == nil {
		return nil, fmt.Errorf("failed to parse Fixer response: %+v", resp)
	}

	results := make([]FetchResult, len(pairs))
	for i, pair := range pairs {
		if price, ok := resp.Rates[f.renameSymbol(pair.Quote)]; ok {
			// Build Price from exchange response.
			results[i] = FetchResult{
				Price: Price{
					Pair:      pair,
					Price:     price,
					Timestamp: time.Now(),
				},
				Error: nil,
			}
		} else {
			// Missing quote in exchange response.
			results[i] = fetchResultWithError(
				pair,
				fmt.Errorf("no price for %s quote exist in response %s", pair.Quote, res.Body),
			)
		}
	}
	return results, nil
}
