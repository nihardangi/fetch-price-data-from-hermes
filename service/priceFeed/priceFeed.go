package priceFeed

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
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
	hexEncodedDataBytes, err := hex.DecodeString(hexEncodedData)
	if err != nil {
		fmt.Println(err)
		return
	}
	var newHexEncodedBytes []byte
	for i := 0; i < len(hexEncodedDataBytes); i++ {
		// 10 byte PNAU VAA wrapper
		pnauVAAWrapperLength := 10
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+pnauVAAWrapperLength]...)
		i += pnauVAAWrapperLength

		// 1 byte version
		versionLength := 1
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+versionLength]...)
		i += versionLength

		// 4 byte guardian_set
		guardianSetLength := 4
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+guardianSetLength]...)
		i += guardianSetLength

		// 1 byte sig_count
		sigCountLength := 1
		buf := bytes.NewReader(hexEncodedDataBytes[i : i+sigCountLength])
		var sigCount uint8
		if err := binary.Read(buf, binary.BigEndian, &sigCount); err != nil {
			fmt.Println(err)
			return
		}
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+sigCountLength]...)
		i += sigCountLength
		fmt.Println("sigCount------", sigCount)

		// Wormhole VAA
		var wormholeVAALength uint
		wormholeVAALength = 66 * uint(sigCount)
		fmt.Println("wormholeVAALen", wormholeVAALength)
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+int(wormholeVAALength)]...)
		i += int(wormholeVAALength)
		fmt.Println("len after-------", len(newHexEncodedBytes))

		// Body Header
		bodyHeaderLength := 51
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+bodyHeaderLength]...)
		i += bodyHeaderLength
		fmt.Println("len after body header-------", len(newHexEncodedBytes))

		// Magic 0x41555756 ("AUWV"). (Mnemonic: Accumulator-Update Wormhole Verification)
		magicBytesLength := 4
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+int(magicBytesLength)]...)
		i += magicBytesLength

		// Major version
		majorBytesLength := 1
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+int(majorBytesLength)]...)
		i += majorBytesLength

		// Minor version
		minorBytesLength := 1
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+int(minorBytesLength)]...)
		i += minorBytesLength

		// Trailing header size
		trailingHeaderTotalBytes := 1
		buf2 := bytes.NewReader(hexEncodedDataBytes[i : i+trailingHeaderTotalBytes])
		var trailingHeaderLength uint8
		if err := binary.Read(buf2, binary.BigEndian, &trailingHeaderLength); err != nil {
			fmt.Println(err)
			return
		}
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+trailingHeaderTotalBytes]...)
		i += trailingHeaderTotalBytes

		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+int(trailingHeaderLength)]...)
		i += int(trailingHeaderLength)
		fmt.Println("trailingHeaderLen", trailingHeaderLength)

		// Proof type
		proofTypeBytesLength := 1
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+proofTypeBytesLength]...)
		fmt.Println("proof type bytes-----------", hexEncodedDataBytes[i:i+proofTypeBytesLength])
		i += proofTypeBytesLength

		// VAA size
		vaaSizeBytesLength := 2
		var vaaSize uint16
		buf3 := bytes.NewReader(hexEncodedDataBytes[i : i+vaaSizeBytesLength])
		fmt.Println("vaa size bytes-----------", hexEncodedDataBytes[i:i+vaaSizeBytesLength])
		if err := binary.Read(buf3, binary.BigEndian, &vaaSize); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("vaa size bytes-----------", hexEncodedDataBytes[i:i+vaaSizeBytesLength])
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+vaaSizeBytesLength]...)
		i += vaaSizeBytesLength
		fmt.Println("vaaSize------", vaaSize)

		// VAA
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+int(vaaSize)]...)
		i += int(vaaSize)

		// Unknown bytes
		unknownBytesLength := 14
		newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+unknownBytesLength]...)
		i += unknownBytesLength

		// Number of updates
		numberOfUpdatesBytesLength := 1
		buf4 := bytes.NewReader(hexEncodedDataBytes[i : i+numberOfUpdatesBytesLength])
		var numberOfUpdates uint8
		if err := binary.Read(buf4, binary.BigEndian, &numberOfUpdates); err != nil {
			fmt.Println(err)
			return
		}
		var newNumberOfUpdates uint8 = uint8(len(priceFeedIDsMap))
		newHexEncodedBytes = append(newHexEncodedBytes, byte(newNumberOfUpdates))
		fmt.Println("new updates", byte(newNumberOfUpdates))
		i += numberOfUpdatesBytesLength

		// Parse each update
		for j := 0; j < int(numberOfUpdates); j++ {
			index := i
			// message size
			messageSizeBytesLength := 2
			buf5 := bytes.NewReader(hexEncodedDataBytes[index : index+messageSizeBytesLength])
			var messageSize uint16
			if err := binary.Read(buf5, binary.BigEndian, &messageSize); err != nil {
				fmt.Println(err)
				return
			}
			// newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:i+messageSizeBytesLength]...)
			index += messageSizeBytesLength
			fmt.Println(messageSize)

			// Extract Price Feed ID
			priceFeedIDBytesLength := 32
			priceFeedID := hex.EncodeToString(hexEncodedDataBytes[index+1 : index+1+priceFeedIDBytesLength])
			fmt.Println("priceFeedID---------", priceFeedID)

			index += int(messageSize)
			numOfProofsBytes := 1
			buf6 := bytes.NewReader(hexEncodedDataBytes[index : index+numOfProofsBytes])
			var numOfProofs uint8
			if err := binary.Read(buf6, binary.BigEndian, &numOfProofs); err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("numOFProofs------", numOfProofs)
			singleProofBytes := 20
			totalProofBytes := singleProofBytes * int(numOfProofs)
			fmt.Println(index)
			index += numOfProofsBytes + totalProofBytes
			fmt.Println("index after----", index)
			// i+=3+messageSize+
			// break
			// continue
			if _, ok := priceFeedIDsMap[fmt.Sprintf("%s%s", "0x", priceFeedID)]; !ok {
				i = index
				fmt.Println("INSIDE MAP ENTRY NOT FOUND CONDITION------")
				continue
			}
			newHexEncodedBytes = append(newHexEncodedBytes, hexEncodedDataBytes[i:index]...)
			priceFeed := hex.EncodeToString(hexEncodedDataBytes[i:index])
			fmt.Println("total price feed data-----", len(priceFeed)/2, priceFeed)
			fmt.Println("last byte", hexEncodedDataBytes[index-1:index])
			i = index
			// fmt.Println("next bytes----", hexEncodedDataBytes[i:i+2])
		}

		// fmt.Println(hexEncodedDataBytes[i : i+5])

	}
	fmt.Println(hex.EncodeToString(newHexEncodedBytes))

}
