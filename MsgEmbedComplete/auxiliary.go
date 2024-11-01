package main

import (
	"covertCommunication/KeyDerivation"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"log"
)

// 生成sourceAddr到destAddr的原始交易（将UTXO全部转给目标地址，没有交易费）
func generateTrans(sourceAddr, destAddr string) *wire.MsgTx {
	// 构造输入
	var inputs []btcjson.TransactionInput
	if len(UTXObyAddress[sourceAddr]) == 0 {
		log.Fatalf("%s have not UTXO", sourceAddr)
	}
	utxo := UTXObyAddress[sourceAddr][0]
	inputs = append(inputs, btcjson.TransactionInput{
		Txid: utxo.TxID,
		Vout: utxo.Vout,
	})
	//	构造输出
	outAddr, _ := btcutil.DecodeAddress(destAddr, &chaincfg.SimNetParams)
	outputs := map[btcutil.Address]btcutil.Amount{
		outAddr: btcutil.Amount(utxo.Amount),
	}
	//	创建交易
	rawTx, err := client.CreateRawTransaction(inputs, outputs, nil)
	if err != nil {
		log.Fatalf("Error creating raw transaction: %v", err)
	}
	return rawTx
}

// 签名交易，嵌入秘密消息
func signTrans(rawTx *wire.MsgTx, embedMsg *string) *wire.MsgTx {
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
func broadTrans(signedTx *wire.MsgTx) string {
	txHash, err := client.SendRawTransaction(signedTx, false)
	if err != nil {
		log.Fatalf("Error sending transaction: %v", err)
	}
	return txHash.String()
}

func importKeyToWallet(key *KeyDerivation.PrivateKey) {
	prikWIF := KeyDerivation.ToWIF(key.Key, "simnet")
	wif, _ := btcutil.DecodeWIF(prikWIF)
	err := client.ImportPrivKey(wif)
	if err != nil {
		log.Fatalf("ImportPrivKey error: %s", err)
	}
}

// 将字符串每32字节划分
func splitStr() []string {
	byteSlice := []byte(Covertmsg)
	chunkSize := 32
	var chunks []string

	for i := 0; i < len(byteSlice); i += chunkSize {
		end := i + chunkSize
		if end > len(byteSlice) {
			end = len(byteSlice)
		}
		chunks = append(chunks, string(byteSlice[i:end]))
	}
	return chunks
}
