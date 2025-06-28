package priceFeed

import (
	"encoding/json"
	"io"
	"net/http"
)

// JSON response returned from the `v2/updates/price/latest` endpoint.
type latestPriceUpdateResponse struct {
	Binary struct {
		Data []string `json:"data"`
	} `json:"binary"`
}

// Unmarshals the response recieved from latest price update API
func handleLatestPriceUpdateResponse(response *http.Response) (latestPriceUpdateResponse, error) {
	data, _ := io.ReadAll(response.Body)
	response.Body.Close()
	var latestPriceUpdateResponseData latestPriceUpdateResponse
	if err := json.Unmarshal(data, &latestPriceUpdateResponseData); err != nil {
		return latestPriceUpdateResponse{}, err
	}
	return latestPriceUpdateResponseData, nil
}
