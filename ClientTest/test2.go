package main

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func main() {
	txid, _ := chainhash.NewHashFromStr("d61eddf4be95e548e23f3ba3b24f4d7aaed098306b6c057ae66f0500f8e6ddc1")
	fmt.Println(txid)
	fmt.Println(txid.String())

}
