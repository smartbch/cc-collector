package main

import (
	"encoding/hex"
	"io"
	"net/http"
)

func getSigByHash(operatorUrl string, txSigHash []byte) ([]byte, error) {
	fullUrl := operatorUrl + "?hash=" + hex.EncodeToString(txSigHash)
	resp, err := http.Get(fullUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
