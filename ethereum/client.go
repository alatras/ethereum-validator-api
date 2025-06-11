package ethereum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	rpcURL     string
	httpClient *http.Client
}

// NewClient creates a new Ethereum RPC client
func NewClient(rpcURL string) (*Client, error) {
	return &Client{
		rpcURL: rpcURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Block represents an Ethereum block
type Block struct {
	Number        string        `json:"number"`
	Hash          string        `json:"hash"`
	BaseFeePerGas string        `json:"baseFeePerGas"`
	GasUsed       string        `json:"gasUsed"`
	Miner         string        `json:"miner"`
	Transactions  []Transaction `json:"transactions"`
}

// Transaction represents an Ethereum transaction
type Transaction struct {
	Hash                 string `json:"hash"`
	GasPrice             string `json:"gasPrice"`
	MaxFeePerGas         string `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas,omitempty"`
	Gas                  string `json:"gas"`
}

// BeaconBlock represents a beacon chain block
type BeaconBlock struct {
	Data struct {
		Message struct {
			Slot          string `json:"slot"`
			ProposerIndex string `json:"proposer_index"`
			Body          struct {
				ExecutionPayload struct {
					FeeRecipient string `json:"fee_recipient"`
					BlockNumber  string `json:"block_number"`
				} `json:"execution_payload"`
			} `json:"body"`
		} `json:"message"`
	} `json:"data"`
}

// ValidatorResponse represents validator data
type ValidatorResponse struct {
	Data struct {
		Validator struct {
			WithdrawalCredentials string `json:"withdrawal_credentials"`
		} `json:"validator"`
	} `json:"data"`
}

// SyncCommittee represents sync committee data
type SyncCommittee struct {
	Data struct {
		Validators []string `json:"validators"`
	} `json:"data"`
}

// GetCurrentSlot retrieves the current slot from the beacon chain
func (c *Client) GetCurrentSlot() (uint64, error) {
	resp, err := c.makeBeaconRequest("GET", "/eth/v1/beacon/headers/head", nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Header struct {
				Message struct {
					Slot string `json:"slot"`
				} `json:"message"`
			} `json:"header"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	slot := new(big.Int)
	slot.SetString(result.Data.Header.Message.Slot, 10)
	return slot.Uint64(), nil
}

// GetBeaconBlock retrieves a beacon block by slot
func (c *Client) GetBeaconBlock(slot uint64) (*BeaconBlock, error) {
	endpoint := fmt.Sprintf("/eth/v2/beacon/blocks/%d", slot)
	resp, err := c.makeBeaconRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("block not found")
	}

	var block BeaconBlock
	if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
		return nil, err
	}

	return &block, nil
}

// GetValidator retrieves validator data by index
func (c *Client) GetValidator(index string) (*ValidatorResponse, error) {
	endpoint := fmt.Sprintf("/eth/v1/beacon/states/head/validators/%s", index)
	resp, err := c.makeBeaconRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var validator ValidatorResponse
	if err := json.NewDecoder(resp.Body).Decode(&validator); err != nil {
		return nil, err
	}

	return &validator, nil
}

// GetExecutionBlock retrieves an execution layer block
func (c *Client) GetExecutionBlock(blockNumber string) (*Block, error) {
	params := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{blockNumber, true},
		"id":      1,
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.rpcURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result *Block `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", result.Error.Message)
	}

	return result.Result, nil
}

// GetSyncCommittee retrieves sync committee for a specific slot
func (c *Client) GetSyncCommittee(slot uint64) (*SyncCommittee, error) {
	// Calculate sync committee period (512 epochs)
	epoch := slot / 32
	syncPeriod := epoch / 256
	stateSlot := syncPeriod * 256 * 32 // First slot of the sync period

	endpoint := fmt.Sprintf("/eth/v1/beacon/states/%d/sync_committees", stateSlot)
	resp, err := c.makeBeaconRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("sync committee not found")
	}

	var committee SyncCommittee
	if err := json.NewDecoder(resp.Body).Decode(&committee); err != nil {
		return nil, err
	}

	return &committee, nil
}

// makeBeaconRequest makes a request to the beacon API
func (c *Client) makeBeaconRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	// Extract base URL and construct beacon API URL
	baseURL := strings.TrimSuffix(c.rpcURL, "/")
	url := baseURL + endpoint

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(req)
}

// CalculateBlockReward calculates the total block reward in GWEI
func CalculateBlockReward(block *Block) (uint64, error) {
	baseFee := new(big.Int)
	baseFee.SetString(block.BaseFeePerGas[2:], 16) // Remove "0x" prefix

	gasUsed := new(big.Int)
	gasUsed.SetString(block.GasUsed[2:], 16)

	// Calculate base fee * gas used
	baseFeeReward := new(big.Int).Mul(baseFee, gasUsed)

	// Calculate total priority fees
	totalPriorityFees := big.NewInt(0)
	for _, tx := range block.Transactions {
		var priorityFee *big.Int

		// EIP-1559 transaction
		if tx.MaxPriorityFeePerGas != "" {
			maxPriorityFee := new(big.Int)
			maxPriorityFee.SetString(tx.MaxPriorityFeePerGas[2:], 16)

			maxFee := new(big.Int)
			maxFee.SetString(tx.MaxFeePerGas[2:], 16)

			// Calculate effective priority fee
			effectiveFee := new(big.Int).Sub(maxFee, baseFee)
			if effectiveFee.Cmp(maxPriorityFee) > 0 {
				priorityFee = maxPriorityFee
			} else {
				priorityFee = effectiveFee
			}
		} else {
			// Legacy transaction
			gasPrice := new(big.Int)
			gasPrice.SetString(tx.GasPrice[2:], 16)
			priorityFee = new(big.Int).Sub(gasPrice, baseFee)
		}

		if priorityFee.Sign() > 0 {
			gas := new(big.Int)
			gas.SetString(tx.Gas[2:], 16)
			txFee := new(big.Int).Mul(priorityFee, gas)
			totalPriorityFees.Add(totalPriorityFees, txFee)
		}
	}

	// Total reward = base fee reward + priority fees
	totalReward := new(big.Int).Add(baseFeeReward, totalPriorityFees)

	// Convert to GWEI (divide by 1e9)
	gwei := new(big.Int).Div(totalReward, big.NewInt(1e9))

	return gwei.Uint64(), nil
}
