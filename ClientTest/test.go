package main

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
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

	fmt.Println(filterTransByInputaddr(client, "SiGGuKwQ2WP1uZ63TBVk1E6mb3qPyqrnEg"))

}

// // filterTrans 根据输入地址筛选交易
func filterTransByInputaddr(client *rpcclient.Client, addr string) []string {
	var txIds []string
	transactions, _ := client.ListTransactionsCount("*", 100)
	// 遍历所有交易依次筛选
	for _, v := range transactions {
		// coinbase交易没有输入
		if v.Generated {
			continue
		}
		// 获取交易的输入utxo
		inputUTXO := getInputUTXO(client, v.TxID)
		// 根据输入utxo提取输入地址
		var inputAddr []string
		for _, utxo := range inputUTXO {
			// 产生这个utxo的交易id
			utxoHash := utxo.Hash
			// utxo在交易中的序号
			utxoIndex := utxo.Index
			previousTrans, _ := client.GetTransaction(&utxoHash)
			var addrTemp string
			// 从details找到对应的vout，（每一个输出都会在details中插入两条记录，一个send类型，一个receive类型，coinbase交易只有一个为generate类型）
			if previousTrans.Details[0].Category == "generate" {
				continue
			}
			addrTemp = previousTrans.Details[2*utxoIndex+1].Address
			inputAddr = append(inputAddr, addrTemp)
			// 交易的输入地址包含目标地址
			if addrTemp == addr {
				txIds = append(txIds, v.TxID)
				break
			}
		}
	}
	return txIds

}

// getInputAddr 获取交易的输入UTXO
func getInputUTXO(client *rpcclient.Client, txid string) []wire.OutPoint {
	// 根据交易id筛选原始交易（原始交易包含txin字段）
	txHash, _ := chainhash.NewHashFromStr(txid)
	rawTransaction, _ := client.GetRawTransaction(txHash)
	txInput := rawTransaction.MsgTx().TxIn
	var inputUtxo []wire.OutPoint
	for _, v := range txInput {
		inputUtxo = append(inputUtxo, v.PreviousOutPoint)
	}
	return inputUtxo
}
