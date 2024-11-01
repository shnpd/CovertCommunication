package main

import (
	"covertCommunication/KeyDerivation"
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
)

// init 生成根密钥对
func init() {
	skroot, _ = KeyDerivation.GenerateMasterKey([]byte("initseed"))
	pkroot = KeyDerivation.EntirePublicKeyForPrivateKey(skroot)
	bankRoot, _ = KeyDerivation.GenerateMasterKey([]byte("bank"))
	kleak = "leak Random"
}

// 初始化钱包
func initWallet() {
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
	client, _ = rpcclient.New(connCfg, nil)
	client.WalletPassphrase("ts0", 6000)

	//	预先生成第一组私钥,预保存5个UTXO
	generateNewCntPrivKeys(5)
}

func main() {
	initWallet()
	defer client.Shutdown()

	//	计算嵌入消息的字节数，每个交易可以嵌入32字节
	byteCnt := len([]byte(Covertmsg))
	msgCnt := (byteCnt + 31) / 32
	// 字符串每32字节划分
	splitMsg := splitStr()

	//	生成msgCnt个私钥用来接收交易
	generateNewCntPrivKeys(msgCnt)

	//通信轮数=一层私钥数-1
	//第i轮通信为i-1到i的交易
	commId = len(layer1Prik) - 1
	updateMapUTXOFromAddr()
	// 使用银行地址集平衡UTXO数量
	transferBank(commId-1, msgCnt)
	updateMapUTXOFromAddr()

	//	发送隐蔽交易
	var covertTxid []string
	covertTxid = sendCovertTransaction(commId-1, msgCnt, splitMsg)
	for _, v := range covertTxid {
		fmt.Printf("Covert transaction id: %s\n", v)
	}

	// 发送泄露交易
	sendLeakTrans(commId)
}

func sendLeakTrans(commId int) {
	sourceAddr := KeyDerivation.GetAddressByPrivKey(layer1Prik[commId])
	rawTx := generateTrans(sourceAddr, "SiGGuKwQ2WP1uZ63TBVk1E6mb3qPyqrnEg")
	signedTx := signTrans(rawTx, &kleak)
	broadTrans(signedTx)
}

// 第commid次通信，处理银行地址，多退少补
func transferBank(commId, msgCnt int) {
	utxoNum := len(prikSet[commId])

	//utxo有多余
	if utxoNum > msgCnt {
		minus := utxoNum - msgCnt
		generateCntBankKeys(minus)
		for i := 0; i < minus; i++ {
			// 创建转出交易
			sourceAddr := addressSet[commId][msgCnt+i]

			destAddr := KeyDerivation.GetAddressByPrivKey(bankPriks[len(bankPriks)-i-1])
			rawTx := generateTrans(sourceAddr, destAddr)
			// 签名交易，该交易不嵌入信息
			signedTx := signTrans(rawTx, nil)
			broadTrans(signedTx)
		}
	} else {
		//	utxo数量不足
		minus := msgCnt - utxoNum
		for i := 0; i < minus; i++ {
			//提取银行地址(默认银行地址足够用)
			bankPrik_t := bankPriks[0]
			bankPriks = bankPriks[1:]
			sourceAddr := KeyDerivation.GetAddressByPrivKey(bankPrik_t)
			// 生成地址接收银行utxo
			addCntPrivKeys(commId, minus)
			destAddr := addressSet[commId][utxoNum+i]
			rawTx := generateTrans(sourceAddr, destAddr)
			signedTx := signTrans(rawTx, nil)
			broadTrans(signedTx)
		}
	}
}

// generateNewCntPrivKey 新生成一个一层私钥，并在这个私钥下派生cnt个私钥，返回一层私钥的序号
func generateNewCntPrivKeys(cnt int) int {
	parentPrik, _ := skroot.ChildPrivateKeyDeprive(uint32(len(layer1Prik)))
	layer1Prik = append(layer1Prik, parentPrik)
	var priksetTemp []*KeyDerivation.PrivateKey
	var addressTemp []string
	for i := 0; i < cnt; i++ {
		key, _ := parentPrik.ChildPrivateKeyDeprive(uint32(i))
		importKeyToWallet(key)
		priksetTemp = append(priksetTemp, key)
		//更新地址集
		addressTemp = append(addressTemp, KeyDerivation.GetAddressByPrivKey(key))
	}
	prikSet = append(prikSet, priksetTemp)
	addressSet = append(addressSet, addressTemp)
	return len(layer1Prik)
}

// addCntPrivKeys 为第i组密钥追加派生cnt个密钥
func addCntPrivKeys(id, cnt int) {
	parentPrik := layer1Prik[id]
	lenPrik := len(prikSet[id])
	for i := 0; i < cnt; i++ {
		key, _ := parentPrik.ChildPrivateKeyDeprive(uint32(lenPrik + i))
		importKeyToWallet(key)
		prikSet[id] = append(prikSet[id], key)
		//更新地址集
		addressSet[id] = append(addressSet[id], KeyDerivation.GetAddressByPrivKey(key))
	}
}

// generateCntBankKeys 生成cnt个银行地址
func generateCntBankKeys(cnt int) {
	parentKey := bankRoot
	for i := 0; i < cnt; i++ {
		key, _ := parentKey.ChildPrivateKeyDeprive(uint32(bankId))
		importKeyToWallet(key)
		bankPriks = append(bankPriks, key)
		bankId++
	}
}

// 更新地址持有UTXO的映射
func updateMapUTXOFromAddr() {
	allUTXO, _ := client.ListUnspent()
	for _, utxo := range allUTXO {
		UTXObyAddress[utxo.Address] = append(UTXObyAddress[utxo.Address], utxo)
	}
}

// 发送隐蔽交易
func sendCovertTransaction(commId, msgCnt int, splitMsg []string) (covertTxid []string) {
	//	构造addrcnt个隐蔽交易,每个交易只有1个输入1个输出
	for i := 0; i < msgCnt; i++ {
		// 创建交易
		rawTx := generateTrans(addressSet[commId][i], addressSet[commId+1][i])
		//	签名交易(嵌入消息)
		signedTx := signTrans(rawTx, &splitMsg[i])
		//	广播交易
		txId := broadTrans(signedTx)
		covertTxid = append(covertTxid, txId)
	}
	return covertTxid
}
