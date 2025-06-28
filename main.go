package main

import (
	"fetch-price-data-from-hermes/service/priceFeed"
	"fmt"
)

func main() {
	// 20 random priceFeed IDs
	priceFeedID1 := "0x7a36855b8a4a6efd701ed82688694bbf67602de9faae509ae28f91065013cb82"
	priceFeedID2 := "0x15add95022ae13563a11992e727c91bdb6b55bc183d9d747436c80a483d8c864"
	priceFeedID3 := "0x95ea50020cf75a81a105d639fd74773ade522e12044600b52286ff5961c71412"
	priceFeedID4 := "0x03ae4db29ed4ae33d323568895aa00337e658e348b37509f5372ae51f0af00d5"
	priceFeedID5 := "0xf610eae82767039ffc95eef8feaeddb7bbac0673cfe7773b2fde24fd1adb0aee"
	priceFeedID6 := "0x3fa4252848f9f0a1480be62745a4629d9eb1322aebab8a791e344b3b9c1adcf5"
	priceFeedID7 := "0xac3d02825eef9c311197bcdd3ef789b2c7db97739af8f6aeca8374c55b022572"
	priceFeedID8 := "0x7677dd124dee46cfcd46ff03cf405fb0ed94b1f49efbea3444aadbda939a7ad3"
	priceFeedID9 := "0x1aef60609549bdab71241b9d031dd2d2573b5a1f591fc87091ab4af67d81e91a"
	priceFeedID10 := "0x89b814de1eb2afd3d3b498d296fca3a873e644bafb587e84d181a01edd682853"
	priceFeedID11 := "0xf6b551a947e7990089e2d5149b1e44b369fcc6ad3627cb822362a2b19d24ad4a"
	priceFeedID12 := "0x681e0eb7acf9a2a3384927684d932560fb6f67c6beb21baa0f110e993b265386"
	priceFeedID13 := "0xb00b60f88b03a6a625a8d1c048c3f66653edf217439983d037e7222c4e612819"
	priceFeedID14 := "0x2ea070725c82f69be1a730c1730cb229dc3ab44459f41d6f06f0b9ab551e4ddb"
	priceFeedID15 := "0x2f7c4f738d498585065a4b87b637069ec99474597da7f0ca349ba8ac3ba9cac5"
	priceFeedID16 := "0xd9912df360b5b7f21a122f15bdd5e27f62ce5e72bd316c291f7c86620e07fb2a"
	priceFeedID17 := "0x37c307959acbb353e1451bcf7da9d305c8cb8d54c64353588aaf900ffcffdd7d"
	priceFeedID18 := "0x93da3352f9f1d105fdfe4971cfa80e9dd777bfc5d0f683ebb6e1294b92137bb7"
	priceFeedID19 := "0x60144b1d5c9e9851732ad1d9760e3485ef80be39b984f6bf60f82b28a2b7f126"
	priceFeedID20 := "0xb7e3904c08ddd9c0c10c6d207d390fd19e87eb6aab96304f571ed94caebdefa0"

	// Fetch data for 20 price feed IDs
	priceFeedIDs := []string{priceFeedID1, priceFeedID2, priceFeedID3, priceFeedID4, priceFeedID5, priceFeedID6, priceFeedID7, priceFeedID8, priceFeedID9, priceFeedID10, priceFeedID11, priceFeedID12, priceFeedID13, priceFeedID14, priceFeedID15, priceFeedID16, priceFeedID17, priceFeedID18, priceFeedID19, priceFeedID20}
	latestPriceFeedData, _ := priceFeed.FetchLatestPriceFeedData(priceFeedIDs)
	fmt.Println("latest feed data----------", latestPriceFeedData)

	// Now, prepare data only for 5 price feed IDs.
	updatePriceFeedIDs := []string{priceFeedID1, priceFeedID3, priceFeedID5, priceFeedID7, priceFeedID9}
	priceFeed.PrepareDataForUpdatePriceFeeds(latestPriceFeedData, updatePriceFeedIDs)
}
