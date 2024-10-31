package main

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"log"
)

const btc = 1000000000

//func main() {
//	initWallet()
//	rawtx := generateTransFromUTXO("3baeb892ff56ef2eef967c77ffc9995e799307e08ed471780d22386ea52a31d2", "STGeYnmKs1XRRUdY5xBWQgkDe12XM69uPR")
//	signtx := signTrans2(rawtx, nil)
//
//	fmt.Println(broadTrans2(signtx))
//}

// 生成sourceAddr到destAddr的原始交易（将UTXO全部转给目标地址，没有交易费）
func generateTransFromUTXO(txid, destAddr string) *wire.MsgTx {
	// 构造输入
	var inputs []btcjson.TransactionInput
	inputs = append(inputs, btcjson.TransactionInput{
		Txid: txid,
		Vout: 0,
	})
	//	构造输出
	outAddr, _ := btcutil.DecodeAddress(destAddr, &chaincfg.SimNetParams)
	outputs := map[btcutil.Address]btcutil.Amount{
		outAddr: btcutil.Amount(5 * btc),
	}
	//	创建交易
	rawTx, err := client.CreateRawTransaction(inputs, outputs, nil)
	if err != nil {
		log.Fatalf("Error creating raw transaction: %v", err)
	}
	return rawTx
}

// 签名交易，嵌入秘密消息
func signTrans2(rawTx *wire.MsgTx, embedMsg *string) *wire.MsgTx {
	signedTx, complete, err := client.SignRawTransaction(rawTx, embedMsg)
	if err != nil {
		log.Fatalf("Error signing transaction: %v", err)
	}
	if !complete {
		log.Fatalf("Transaction signing incomplete")
	}
	return signedTx
}

// 广播交易
func broadTrans2(signedTx *wire.MsgTx) string {
	txHash, err := client.SendRawTransaction(signedTx, false)
	if err != nil {
		log.Fatalf("Error sending transaction: %v", err)
	}
	return txHash.String()
}