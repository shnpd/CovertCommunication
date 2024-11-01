package main

import (
	"covertCommunication/KeyDerivation"
	"crypto/sha256"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"log"
)

var t int64
var val float64

func main() {
	seed := []byte("111")
	sk, _ := KeyDerivation.GenerateMasterKey(seed)
	pk := KeyDerivation.EntirePublicKeyForPrivateKey(sk)
	//pubKey, err := secp256k1.ParsePubKey(pk.Key)
	//if err != nil {
	//	return
	//}
	//addr, err := btcutil.NewAddressPubKey(pubKey.SerializeUncompressed(), &chaincfg.SimNetParams)
	//fmt.Println(addr.EncodeAddress())
	//fmt.Println(KeyDerivation.GetAddressByPrivKey(sk))
	//fmt.Println(GetAddressByPubKey(pk))
	fmt.Println(GetAddressByPrivKey(sk))
	fmt.Println(GetAddressByPubKey(pk))
}

// 根据PublicKey获取对应的地址
func GetAddressByPubKey(key *KeyDerivation.PublicKey) string {
	// 生成地址
	pubKey, err := secp256k1.ParsePubKey(key.Key)
	if err != nil {
		log.Fatalf("get address error %s", err)
	}
	addr, err := btcutil.NewAddressPubKey(pubKey.SerializeUncompressed(), &chaincfg.SimNetParams)
	return addr.EncodeAddress()
}

// 根据wif私钥获取对应的地址
func GetAddressByWIF(keywif string) (string, error) {
	// 解析WIF格式
	privKey, err := btcutil.DecodeWIF(keywif)
	if err != nil {
		fmt.Println("Error decoding WIF:", err)
		return "", err
	}
	// 计算公钥
	pubKey := privKey.PrivKey.PubKey()
	// 生成地址
	addr, err := btcutil.NewAddressPubKey(pubKey.SerializeUncompressed(), &chaincfg.SimNetParams)
	if err != nil {
		fmt.Println("Error creating address:", err)
		return "", err
	}
	// 输出地址
	return addr.EncodeAddress(), nil
}

// 根据PrivateKey获取对应的地址
func GetAddressByPrivKey(key *KeyDerivation.PrivateKey) string {
	prikWIF := ToWIF(key.Key, "simnet")
	address, _ := GetAddressByWIF(prikWIF)
	return address
}

// ToWIF 将私钥转换为 WIF 格式
func ToWIF(privateKey []byte, netType string) string {
	// 选择网络字节
	var networkByte byte
	switch netType {
	case "mainnet":
		networkByte = 0x80 // 主网
	case "testnet":
		networkByte = 0xEF // 测试网
	case "simnet":
		networkByte = 0x64 //模拟网
	default:
		log.Fatalf("netType error")
	}

	// 创建新的字节数组，长度为私钥长度 + 1 + 4（校验和）
	wif := make([]byte, 0, len(privateKey)+1+4)
	wif = append(wif, networkByte)   // 添加网络字节
	wif = append(wif, privateKey...) // 添加私钥

	// 计算校验和
	checksum := sha256.Sum256(wif)        // 第一次SHA-256
	checksum = sha256.Sum256(checksum[:]) // 第二次SHA-256
	checksum2 := checksum[:4]             // 取前4字节作为校验和

	// 将校验和添加到WIF末尾
	wif = append(wif, checksum2...)

	// 进行Base58编码
	return base58.Encode(wif)
}
