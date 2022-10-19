package collector

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	gethcmn "github.com/ethereum/go-ethereum/common"
)

type OperatorResp struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Result  string `json:"result,omitempty"`
}

func getSigByHash(operatorUrl string, txSigHash []byte) ([]byte, error) {
	fullUrl := operatorUrl + "/sig?hash=" + hex.EncodeToString(txSigHash)
	fmt.Println("getSigByHash:", fullUrl)
	resp, err := http.Get(fullUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respJson OperatorResp
	err = json.Unmarshal(respBytes, &respJson)
	if err != nil {
		return nil, err
	}
	if respJson.Error != "" {
		return nil, errors.New(respJson.Error)
	}

	return gethcmn.FromHex(respJson.Result), nil
}
