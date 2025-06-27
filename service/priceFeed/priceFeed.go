package priceFeed

import (
	"fmt"
	"net/http"
	"time"
)

const hermesLatestPriceUpdateUrl = "https://hermes.pyth.network/v2/updates/price/latest"

var (
	httpClient = http.Client{
		Timeout: 5 * time.Second,
	}
)

func FetchLatestPriceFeedData(priceFeedIDs []string) {
	// create a request object
	req, _ := http.NewRequest(
		"GET",
		hermesLatestPriceUpdateUrl,
		nil,
	)
	q := req.URL.Query()
	for _, priceFeedID := range priceFeedIDs {
		q.Add("ids[]", priceFeedID)
	}
	req.URL.RawQuery = q.Encode()
	fmt.Println("url-----", req.URL.String())
	response, err := httpClient.Do(req)
	if err != nil {

	}
	if response.StatusCode != http.StatusOK {
		fmt.Println("Status code------", response.StatusCode)
		return
	}
	latestPriceUpdateResponse, err := handleLatestPriceUpdateResponse(response)
	if err != nil {
		fmt.Println("handling response error-----", err)
		return
	}
	fmt.Println(latestPriceUpdateResponse)
}

func PrepareDataForUpdatePriceFeeds(priceFeedIDs []string) {
	priceFeedIDsMap := make(map[string]bool, len(priceFeedIDs))
	for _, priceFeedID := range priceFeedIDs {
		priceFeedIDsMap[priceFeedID] = true
	}
}
