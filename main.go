package main

import (
	"fetch-price-data-from-hermes/service/priceFeed"
	"fmt"
)

func main() {
	priceFeedIDs := []string{"0xe62df6c8b4a85fe1a67db44dc12de5db330f7ac66b72dc658afedf0f4a415b43", "0xc96458d393fe9deb7a7d63a0ac41e2898a67a7750dbd166673279e06c868df0a"}
	latestPriceFeedData, _ := priceFeed.FetchLatestPriceFeedData(priceFeedIDs)
	fmt.Println("latest feed data----------", latestPriceFeedData)

	updatePriceFeedIDs := []string{"0xe62df6c8b4a85fe1a67db44dc12de5db330f7ac66b72dc658afedf0f4a415b43"}
	priceFeed.PrepareDataForUpdatePriceFeeds(latestPriceFeedData, updatePriceFeedIDs)
}
