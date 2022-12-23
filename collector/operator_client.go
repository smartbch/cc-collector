package collector

import (
	"time"

	opclient "github.com/smartbch/cc-operator/client"
)

func getSigByHash(operatorUrl string, txSigHash []byte) ([]byte, error) {
	client := opclient.NewClient(operatorUrl, 5*time.Second)
	return client.GetSig(txSigHash)
}
