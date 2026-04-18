package chains

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"
)

var baseClient = http11Client(10 * time.Second)

func Base() (int, error) {
	url := "https://mainnet.base.org"
	payload := []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`)

	var resp *http.Response
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		req, reqErr := http.NewRequest("POST", url, bytes.NewBuffer(payload))
		if reqErr != nil {
			return 0, fmt.Errorf("failed to create base request: %v", reqErr)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) Chainflow/nc-calc")

		resp, err = baseClient.Do(req)
		if err == nil {
			break
		}
		log.Printf("base: request attempt %d failed: %v", attempt+1, err)
		time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
	}
	if err != nil {
		return 0, fmt.Errorf("base rpc unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("base rpc returned status: %d", resp.StatusCode)
	}

	return 1, nil
}