package main

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"log"
)

func main() {
	// 设置RPC客户端连接的配置
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:28335", // 替换为你的btcwallet的RPC地址
		User:         "simnet",          // 在btcwallet配置文件中定义的RPC用户名
		Pass:         "simnet",          // 在btcwallet配置文件中定义的RPC密码
		HTTPPostMode: true,              // 使用HTTP POST模式
		DisableTLS:   true,              // 禁用TLS
		Params:       "simnet",          // 连接到simnet网
	}

	// 创建新的RPC客户端
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatalf("Error creating new client: %v", err)
	}
	defer client.Shutdown()

	//构造输入
	var inputs []btcjson.TransactionInput
	inputs = append(inputs, btcjson.TransactionInput{
		Txid: "f52b43f10c97280e67c0f2e42631f9bcf621efec2d9ed0c9fe1e0784d02359c1",
		Vout: 0,
	})
	var totalInput btcutil.Amount
	totalInput = btcutil.Amount(1)

	//构造输出（输入全部用作输出，没有交易费）
	destinationAddress := "SStEkSyJbAHXdcrfb4GQiLmkS7pEGLDW2Q"
	destAddr, _ := btcutil.DecodeAddress(destinationAddress, &chaincfg.SimNetParams)
	outputs := map[btcutil.Address]btcutil.Amount{
		destAddr: totalInput,
	}

	// 创建交易
	rawTx, err := client.CreateRawTransaction(inputs, outputs, nil)

	if err != nil {
		log.Fatalf("Error creating raw transaction: %v", err)
	}

	//	签名交易(嵌入消息)
	coverMsg := "hello world!"
	signedTx, complete, err := client.SignRawTransaction(rawTx, &coverMsg)
	if err != nil {
		log.Fatalf("Error signing transaction: %v", err)
	}
	if !complete {
		log.Fatalf("Transaction signing incomplete")
	}

	//	广播交易
	txHash, err := client.SendRawTransaction(signedTx, false)
	if err != nil {
		log.Fatalf("Error sending transaction: %v", err)
	}

	log.Printf("Transaction sent successfully: %v", txHash)
}
