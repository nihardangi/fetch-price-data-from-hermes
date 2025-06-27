package priceFeed

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

const (
	hermesLatestPriceUpdateUrl = "https://hermes.pyth.network/v2/updates/price/latest"
	pnauVAAWrapperLength       = 10
	bodyHeaderLength           = 51
)

var (
	httpClient = http.Client{
		Timeout: 5 * time.Second,
	}
)

func FetchLatestPriceFeedData(priceFeedIDs []string) (string, error) {
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
		// Add error
		return "", nil
	}
	latestPriceUpdateResponse, err := handleLatestPriceUpdateResponse(response)
	if err != nil {
		fmt.Println("handling response error-----", err)
		return "", err
	}
	if len(latestPriceUpdateResponse.Binary.Data) == 0 {
		// Add error here
		return "", err
	}
	return latestPriceUpdateResponse.Binary.Data[0], nil
}

func PrepareDataForUpdatePriceFeeds(hexEncodedData string, priceFeedIDs []string) {
	priceFeedIDsMap := make(map[string]bool, len(priceFeedIDs))
	for _, priceFeedID := range priceFeedIDs {
		priceFeedIDsMap[priceFeedID] = true
	}
	originalDataBytes, err := hex.DecodeString(hexEncodedData)
	if err != nil {
		fmt.Println(err)
		return
	}
	var newDataBytes []byte
	for i := 0; i < len(originalDataBytes); i++ {
		// New logic
		cursor := 0
		newDataBytes, cursor = extractAndAddPNAUWrapper(cursor, originalDataBytes, newDataBytes)
		newDataBytes, cursor = extractAndAddWormholeVAA(cursor, originalDataBytes, newDataBytes)
		newDataBytes, cursor = extractAndAddBodyHeader(cursor, originalDataBytes, newDataBytes)

		i = cursor

		// Magic 0x41555756 ("AUWV"). (Mnemonic: Accumulator-Update Wormhole Verification)
		magicBytesLength := 4
		newDataBytes = append(newDataBytes, originalDataBytes[i:i+int(magicBytesLength)]...)
		i += magicBytesLength

		// Major version
		majorBytesLength := 1
		newDataBytes = append(newDataBytes, originalDataBytes[i:i+int(majorBytesLength)]...)
		i += majorBytesLength

		// Minor version
		minorBytesLength := 1
		newDataBytes = append(newDataBytes, originalDataBytes[i:i+int(minorBytesLength)]...)
		i += minorBytesLength

		// Trailing header size
		trailingHeaderTotalBytes := 1
		buf2 := bytes.NewReader(originalDataBytes[i : i+trailingHeaderTotalBytes])
		var trailingHeaderLength uint8
		if err := binary.Read(buf2, binary.BigEndian, &trailingHeaderLength); err != nil {
			fmt.Println(err)
			return
		}
		newDataBytes = append(newDataBytes, originalDataBytes[i:i+trailingHeaderTotalBytes]...)
		i += trailingHeaderTotalBytes

		newDataBytes = append(newDataBytes, originalDataBytes[i:i+int(trailingHeaderLength)]...)
		i += int(trailingHeaderLength)
		fmt.Println("trailingHeaderLen", trailingHeaderLength)

		// Proof type
		proofTypeBytesLength := 1
		newDataBytes = append(newDataBytes, originalDataBytes[i:i+proofTypeBytesLength]...)
		fmt.Println("proof type bytes-----------", originalDataBytes[i:i+proofTypeBytesLength])
		i += proofTypeBytesLength

		newDataBytes, i = extractAndAddVAASizeAndVAA(i, originalDataBytes, newDataBytes)

		// Unknown bytes
		unknownBytesLength := 14
		newDataBytes = append(newDataBytes, originalDataBytes[i:i+unknownBytesLength]...)
		i += unknownBytesLength

		// Number of updates
		numberOfUpdatesBytesLength := 1
		buf4 := bytes.NewReader(originalDataBytes[i : i+numberOfUpdatesBytesLength])
		var numberOfUpdates uint8
		if err := binary.Read(buf4, binary.BigEndian, &numberOfUpdates); err != nil {
			fmt.Println(err)
			return
		}
		var newNumberOfUpdates uint8 = uint8(len(priceFeedIDsMap))
		newDataBytes = append(newDataBytes, byte(newNumberOfUpdates))
		fmt.Println("new updates", byte(newNumberOfUpdates))
		i += numberOfUpdatesBytesLength

		// Parse each update
		cursor = i
		for j := 0; j < int(numberOfUpdates); j++ {
			newDataBytes, cursor = extractAndAddPriceFeedDataIfPresentInMap(cursor, priceFeedIDsMap, originalDataBytes, newDataBytes)
			fmt.Println("after body header-------", hex.EncodeToString(newDataBytes))
			// break
			i = cursor
			// fmt.Println("next bytes----", originalDataBytes[i:i+2])
		}
		// break
		// i = cursor

		// fmt.Println(originalDataBytes[i : i+5])

	}
	// fmt.Println(hex.EncodeToString(newDataBytes))

}

func extractAndAddPNAUWrapper(i int, originalDataBytes, newDataBytes []byte) ([]byte, int) {
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+pnauVAAWrapperLength]...)
	i += pnauVAAWrapperLength
	return newDataBytes, pnauVAAWrapperLength
}

func extractAndAddWormholeVAA(i int, originalDataBytes, newDataBytes []byte) ([]byte, int) {
	// 1 byte version
	versionLength := 1
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+versionLength]...)
	i += versionLength

	// 4 byte guardian_set
	guardianSetLength := 4
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+guardianSetLength]...)
	i += guardianSetLength

	// 1 byte sig_count
	sigCountLength := 1
	buf := bytes.NewReader(originalDataBytes[i : i+sigCountLength])
	var sigCount uint8
	if err := binary.Read(buf, binary.BigEndian, &sigCount); err != nil {
		// Add proper error handling
		fmt.Println(err)
		return []byte{}, 0
	}
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+sigCountLength]...)
	i += sigCountLength
	fmt.Println("sigCount------", sigCount)

	// Wormhole VAA
	var wormholeVAALength uint
	wormholeVAALength = 66 * uint(sigCount)
	fmt.Println("wormholeVAALen", wormholeVAALength)
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+int(wormholeVAALength)]...)
	i += int(wormholeVAALength)
	fmt.Println("len after-------", len(newDataBytes))

	return newDataBytes, i
}

func extractAndAddBodyHeader(i int, originalDataBytes, newDataBytes []byte) ([]byte, int) {
	bodyHeaderLength := 51
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+bodyHeaderLength]...)
	i += bodyHeaderLength
	// fmt.Println("len after body header-------", len(newDataBytes))
	return newDataBytes, i
}

// func extractAccumulatorUpdate

func extractAndAddVAASizeAndVAA(i int, originalDataBytes, newDataBytes []byte) ([]byte, int) {
	// Extract and add VAA size
	vaaSizeBytesLength := 2
	var vaaSize uint16
	buf3 := bytes.NewReader(originalDataBytes[i : i+vaaSizeBytesLength])
	if err := binary.Read(buf3, binary.BigEndian, &vaaSize); err != nil {
		fmt.Println(err)
		// Add proper error handling
		return []byte{}, 0
	}
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+vaaSizeBytesLength]...)
	i += vaaSizeBytesLength
	fmt.Println("vaaSize------", vaaSize)

	// Extract and add VAA
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+int(vaaSize)]...)
	i += int(vaaSize)
	return newDataBytes, i
}

func extractAndAddPriceFeedDataIfPresentInMap(i int, priceFeedIDsMap map[string]bool, originalDataBytes, newDataBytes []byte) ([]byte, int) {
	startIndex := i
	// message size
	messageSizeBytes := 2
	buf1 := bytes.NewReader(originalDataBytes[i : i+messageSizeBytes])
	var messageSize uint16
	if err := binary.Read(buf1, binary.BigEndian, &messageSize); err != nil {
		fmt.Println(err)
		// Add proper error handling
		return []byte{}, 0
	}
	i += messageSizeBytes
	fmt.Println(messageSize)

	// Extract Price Feed ID
	priceFeedIDBytes := 32
	priceFeedID := hex.EncodeToString(originalDataBytes[i+1 : i+1+priceFeedIDBytes])
	fmt.Println("priceFeedID---------", priceFeedID)

	i += int(messageSize)
	numOfProofsBytes := 1
	buf2 := bytes.NewReader(originalDataBytes[i : i+numOfProofsBytes])
	var numOfProofs uint8
	if err := binary.Read(buf2, binary.BigEndian, &numOfProofs); err != nil {
		fmt.Println(err)
		// Add proper error handling
		return []byte{}, 0
	}
	fmt.Println("numOFProofs------", numOfProofs)
	singleProofBytes := 20
	totalProofBytes := singleProofBytes * int(numOfProofs)
	fmt.Println(i)
	i += numOfProofsBytes + totalProofBytes
	fmt.Println("i after----", i)
	// i+=3+messageSize+
	// break
	// continue
	if _, ok := priceFeedIDsMap[fmt.Sprintf("%s%s", "0x", priceFeedID)]; !ok {
		// fmt.Println("INSIDE MAP ENTRY NOT FOUND CONDITION------")
		return newDataBytes, i
	}
	newDataBytes = append(newDataBytes, originalDataBytes[startIndex:i]...)
	// priceFeed := hex.EncodeToString(originalDataBytes[i:i])
	// fmt.Println("total price feed data-----", len(priceFeed)/2, priceFeed)
	// fmt.Println("last byte", originalDataBytes[i-1:i])
	return newDataBytes, i
	// fmt.Println("next bytes----", originalDataBytes[i:i+2])
}
