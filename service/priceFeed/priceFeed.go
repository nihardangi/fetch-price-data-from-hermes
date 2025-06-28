package priceFeed

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	hermesLatestPriceUpdateUrl = "https://hermes.pyth.network/v2/updates/price/latest"
	pnauVAAWrapperLength       = 10
	guardianSetBytes           = 4
	sigCountBytes              = 1
	bytesPerSignature          = 66
	bodyHeaderBytes            = 51
	vaaSizeBytes               = 2
	majorBytes                 = 1
	minorBytes                 = 1
	trailingHeaderTotalBytes   = 1
	proofTypeBytes             = 1
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
	// Prepare URL
	q := req.URL.Query()
	for _, priceFeedID := range priceFeedIDs {
		q.Add("ids[]", priceFeedID)
	}
	req.URL.RawQuery = q.Encode()
	// Send request
	response, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	if response.StatusCode != http.StatusOK {
		err := fmt.Errorf("status code not 200, received status code: %d", response.StatusCode)
		return "", err
	}
	// Parse the response received from API
	latestPriceUpdateResponse, err := handleLatestPriceUpdateResponse(response)
	if err != nil {
		return "", err
	}
	if len(latestPriceUpdateResponse.Binary.Data) == 0 {
		err := errors.New("no data received in API repsonse")
		return "", err
	}
	return latestPriceUpdateResponse.Binary.Data[0], nil
}

func PrepareDataForUpdatePriceFeeds(hexEncodedData string, priceFeedIDs []string) (string, error) {
	// Add priceFeedIDs to a map for uniqueness and faster search
	priceFeedIDsMap := make(map[string]bool, len(priceFeedIDs))
	for _, priceFeedID := range priceFeedIDs {
		priceFeedIDsMap[priceFeedID] = true
	}
	// Convert hex string to bytes
	originalDataBytes, err := hex.DecodeString(hexEncodedData)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	var newDataBytes []byte
	cursor := 0
	newDataBytes, cursor = extractAndAddPNAUWrapper(cursor, originalDataBytes, newDataBytes)
	newDataBytes, cursor = extractAndAddWormholeVAA(cursor, originalDataBytes, newDataBytes)
	newDataBytes, cursor = extractAndAddBodyHeader(cursor, originalDataBytes, newDataBytes)
	newDataBytes, cursor = extractAndAddAccumulatorUpdateHeader(cursor, originalDataBytes, newDataBytes)
	newDataBytes, cursor = extractAndAddVAASizeAndVAA(cursor, originalDataBytes, newDataBytes)

	// Extract and add unknown bytes
	unknownBytesLength := 14
	newDataBytes = append(newDataBytes, originalDataBytes[cursor:cursor+unknownBytesLength]...)
	cursor += unknownBytesLength

	// Extract Number of updates
	numberOfUpdatesBytesLength := 1
	buf4 := bytes.NewReader(originalDataBytes[cursor : cursor+numberOfUpdatesBytesLength])
	var numberOfUpdates uint8
	if err := binary.Read(buf4, binary.BigEndian, &numberOfUpdates); err != nil {
		fmt.Println(err)
		return "", err
	}
	// Add new number of updates i.e total entries in our map.
	var newNumberOfUpdates uint8 = uint8(len(priceFeedIDsMap))
	newDataBytes = append(newDataBytes, byte(newNumberOfUpdates))
	cursor += numberOfUpdatesBytesLength

	// Parse each update
	for j := 0; j < int(numberOfUpdates); j++ {
		newDataBytes, cursor = extractAndAddPriceFeedDataIfPresentInMap(cursor, priceFeedIDsMap, originalDataBytes, newDataBytes)
	}
	fmt.Printf("\nnewHexData for %d priceFeeds is----------------------------\n%s", newNumberOfUpdates, hex.EncodeToString(newDataBytes))
	return hex.EncodeToString(newDataBytes), nil
}

// Extracts and adds 10 byte long PNAU wrapper.
func extractAndAddPNAUWrapper(i int, originalDataBytes, newDataBytes []byte) ([]byte, int) {
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+pnauVAAWrapperLength]...)
	i += pnauVAAWrapperLength
	return newDataBytes, pnauVAAWrapperLength
}

// Extracts and add wormhole VAA fields:
// 1. 1 byte version length
// 2. 4 byte guardian set
// 3. 1 byte sig count
// 4. Signatures usually 66*total signers where bytes in a signature is 66 and total signers is mostly 13
func extractAndAddWormholeVAA(i int, originalDataBytes, newDataBytes []byte) ([]byte, int) {
	// 1 byte version
	versionLength := 1
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+versionLength]...)
	i += versionLength

	// 4 byte guardian_set
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+guardianSetBytes]...)
	i += guardianSetBytes

	// 1 byte sig_count
	buf := bytes.NewReader(originalDataBytes[i : i+sigCountBytes])
	var sigCount uint8
	if err := binary.Read(buf, binary.BigEndian, &sigCount); err != nil {
		// Add proper error handling
		fmt.Println(err)
		return []byte{}, 0
	}
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+sigCountBytes]...)
	i += sigCountBytes

	// Wormhole VAA
	var wormholeVAALength uint
	wormholeVAALength = bytesPerSignature * uint(sigCount)
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+int(wormholeVAALength)]...)
	i += int(wormholeVAALength)

	return newDataBytes, i
}

// Extracts and add body header which is 51 bytes long
func extractAndAddBodyHeader(i int, originalDataBytes, newDataBytes []byte) ([]byte, int) {
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+bodyHeaderBytes]...)
	i += bodyHeaderBytes
	return newDataBytes, i
}

// Extracts and add fields in accumulator update header
// 1. Magic bytes AUWV (4 bytes long)
// 2. 1 byte long Major bytes
// 3. 1 byte long minor bytes
// 4. 1 byte long trailingHeaderSize
// 5. trailingHeaderSize bytes long trailingHeader (usually 0 bytes long)
// 6. 1 byte long proof type
func extractAndAddAccumulatorUpdateHeader(i int, originalDataBytes, newDataBytes []byte) ([]byte, int) {
	// Magic 0x41555756 ("AUWV"). (Mnemonic: Accumulator-Update Wormhole Verification)
	magicBytesLength := 4
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+int(magicBytesLength)]...)
	i += magicBytesLength

	// Major version
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+int(majorBytes)]...)
	i += majorBytes

	// Minor version
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+int(minorBytes)]...)
	i += minorBytes

	// Trailing header size
	buf2 := bytes.NewReader(originalDataBytes[i : i+trailingHeaderTotalBytes])
	var trailingHeaderLength uint8
	if err := binary.Read(buf2, binary.BigEndian, &trailingHeaderLength); err != nil {
		fmt.Println(err)
		// Add Proper error handling
		return []byte{}, 0
	}
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+trailingHeaderTotalBytes]...)
	i += trailingHeaderTotalBytes

	newDataBytes = append(newDataBytes, originalDataBytes[i:i+int(trailingHeaderLength)]...)
	i += int(trailingHeaderLength)

	// Proof type
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+proofTypeBytes]...)
	i += proofTypeBytes
	return newDataBytes, i
}

// Extracts and add VAA size and VAA
// 1. 2 byte long VAA size
// 2. VAA size bytes long VAA
func extractAndAddVAASizeAndVAA(i int, originalDataBytes, newDataBytes []byte) ([]byte, int) {
	// Extract and add VAA size
	var vaaSize uint16
	buf3 := bytes.NewReader(originalDataBytes[i : i+vaaSizeBytes])
	if err := binary.Read(buf3, binary.BigEndian, &vaaSize); err != nil {
		fmt.Println(err)
		// Add proper error handling
		return []byte{}, 0
	}
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+vaaSizeBytes]...)
	i += vaaSizeBytes

	// Extract and add VAA
	newDataBytes = append(newDataBytes, originalDataBytes[i:i+int(vaaSize)]...)
	i += int(vaaSize)
	return newDataBytes, i
}

// Extracts and add price feed data
//  1. Checks if price feed ID is present in map
//     1.1 If present, add message size, price feed update data and all the proofs newDataBytes
//     1.2 If not, skip everything related to priceFeedID i.e message size, price feed update data and all the proofs newDataBytes
//
// Message size is 2 bytes long, usually value is 85
// PriceFeedUpdate Data's length will be equal to message size, usually 85
// Num of proofs is 1 byte long
// After num of proofs field, every proof will be present. Each proof is 20 bytes long.
// For example: if num of proofs is 12, then total proof bytes will be 12*20= 240
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

	// Extract Price Feed ID
	priceFeedIDBytes := 32
	// Adding 1 because in order to fetch priceFeedID, we have to skip the q byte long version
	priceFeedID := hex.EncodeToString(originalDataBytes[i+1 : i+1+priceFeedIDBytes])

	i += int(messageSize)
	// Extract num of proofs
	numOfProofsBytes := 1
	buf2 := bytes.NewReader(originalDataBytes[i : i+numOfProofsBytes])
	var numOfProofs uint8
	if err := binary.Read(buf2, binary.BigEndian, &numOfProofs); err != nil {
		fmt.Println(err)
		// Add proper error handling
		return []byte{}, 0
	}

	singleProofBytes := 20
	totalProofBytes := singleProofBytes * int(numOfProofs)
	i += numOfProofsBytes + totalProofBytes

	priceFeedIDFormatted := fmt.Sprintf("%s%s", "0x", priceFeedID)
	if _, ok := priceFeedIDsMap[priceFeedIDFormatted]; !ok {
		return newDataBytes, i
	}
	newDataBytes = append(newDataBytes, originalDataBytes[startIndex:i]...)
	// Delete the ID from map as it has been added to the packet.
	delete(priceFeedIDsMap, priceFeedIDFormatted)
	return newDataBytes, i
}
