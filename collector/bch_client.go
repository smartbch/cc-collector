package collector

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// https://docs.bitcoincashnode.org/doc/json-rpc/sendrawtransaction/
const (
	SendRawTxReq = `{ "jsonrpc":"1.0", "id":"cc-collector", "method":"sendrawtransaction", "params":["%s"] }`
)

type JsonRpcResp struct {
	Result any           `json:"result"`
	Error  *JsonRpcError `json:"error"`
	Id     string        `json:"id"`
}
type JsonRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type BchRpcClient struct {
	rpcUrl   string
	username string
	password string
}

func newBchClient(rpcUrl, username, password string) *BchRpcClient {
	return &BchRpcClient{
		rpcUrl:   rpcUrl,
		username: username,
		password: password,
	}
}

func (client *BchRpcClient) sendRawTx(txData []byte) error {
	reqStr := fmt.Sprintf(SendRawTxReq, hex.EncodeToString(txData))
	log.Info("sendRawTx req:", reqStr)

	respData, err := client.sendRequest(reqStr)
	if err != nil {
		return err
	}
	log.Info("sendRawTx resp:", string(respData))

	var result JsonRpcResp
	err = json.Unmarshal(respData, &result)
	if err != nil {
		return err
	}

	if result.Error != nil {
		return fmt.Errorf("error code: %d, message: %s",
			result.Error.Code, result.Error.Message)
	}

	return nil
}

func (client *BchRpcClient) sendRequest(reqStr string) ([]byte, error) {
	body := strings.NewReader(reqStr)
	req, err := http.NewRequest("POST", client.rpcUrl, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(client.username, client.password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return respData, nil
}
