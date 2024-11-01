package KeyDerivation

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/ripemd160"
	"io"
	"log"
	"math/big"
)

// point(p):返回secp256k1基点与整数p进行EC点乘法(重复应用EC群运算)得到的坐标对。
func point(p []byte) (*big.Int, *big.Int) {
	return curve.ScalarBaseMult(p)
}

// 将32位无符号整数i序列化为4字节序列，最高有效字节优先
func serializeUint32(i uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, i)
	return buf
}

// serP(P):将坐标对P = (x,y)序列化为字节序列，使用SEC1的压缩形式:(0x02或0x03) || ser256(x)，其中头字节取决于省略的y坐标的奇偶校验。
func serializePub(x *big.Int, y *big.Int) []byte {
	var key bytes.Buffer

	// Write header; 0x2 for even y value; 0x3 for odd
	err := key.WriteByte(byte(0x2) + byte(y.Bit(0)))
	if err != nil {
		return nil
	}

	// Write X coord; Pad the key so x is aligned with the LSB. Pad size is key length - header size (1) - xBytes size
	xBytes := x.Bytes()
	for i := 0; i < (33 - 1 - len(xBytes)); i++ {
		err := key.WriteByte(0x0)
		if err != nil {
			return nil
		}
	}
	_, err = key.Write(xBytes)
	if err != nil {
		return nil
	}
	return key.Bytes()
}

// 将序列化公钥扩展为坐标对形式
func deserializePub(key []byte) (*big.Int, *big.Int) {
	params := curveParams
	exp := big.NewInt(1)
	exp.Add(params.P, exp)
	exp.Div(exp, big.NewInt(4))
	x := big.NewInt(0).SetBytes(key[1:33])
	y := big.NewInt(0).SetBytes(key[:1])
	beta := big.NewInt(0)
	// #nosec
	beta.Exp(x, big.NewInt(3), nil)
	beta.Add(beta, big.NewInt(7))
	// #nosec
	beta.Exp(beta, exp, params.P)
	if y.Add(beta, y).Mod(y, big.NewInt(2)).Int64() == 0 {
		y = beta
	} else {
		y = beta.Sub(params.P, beta)
	}
	return x, y
}

func hashSha256(data []byte) []byte {
	hasher := sha256.New()
	_, err := hasher.Write(data)
	if err != nil {
		return nil
	}
	return hasher.Sum(nil)
}

func hashDoubleSha256(data []byte) []byte {
	return hashSha256(hashSha256(data))
}

func hashRipeMD160(data []byte) []byte {
	hasher := ripemd160.New()
	_, err := io.WriteString(hasher, string(data))
	if err != nil {
		return nil
	}
	return hasher.Sum(nil)
}
func hash160(data []byte) []byte {
	return hashRipeMD160(hashSha256(data))
}

func addPrivateKeys(key1 []byte, key2 []byte) []byte {
	var key1Int big.Int
	var key2Int big.Int
	key1Int.SetBytes(key1)
	key2Int.SetBytes(key2)

	key1Int.Add(&key1Int, &key2Int)
	key1Int.Mod(&key1Int, curve.Params().N)

	b := key1Int.Bytes()
	if len(b) < 32 {
		extra := make([]byte, 32-len(b))
		b = append(extra, b...)
	}
	return b
}
func addPublicKeys(key1 []byte, key2 []byte) []byte {
	x1, y1 := deserializePub(key1)
	x2, y2 := deserializePub(key2)
	return serializePub(curve.Add(x1, y1, x2, y2))
}
func addChecksumToBytes(data []byte) []byte {
	checksum := checksum(data)
	return append(data, checksum...)
}
func checksum(data []byte) []byte {
	return hashDoubleSha256(data)[:4]
}
func PublicKeyForPrivateKey(key []byte) []byte {
	return serializePub(point(key))
}

func EntirePublicKeyForPrivateKey(priv *PrivateKey) *PublicKey {
	ret := &PublicKey{
		Version:      testPublicWalletVersion,
		Depth:        priv.Depth,
		FatherFinger: priv.FatherFinger,
		ChildNumber:  priv.ChildNumber,
		Chaincode:    priv.Chaincode,
		Key:          PublicKeyForPrivateKey(priv.Key),
	}
	return ret
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
func GetAddressByPrivKey(key *PrivateKey) string {
	prikWIF := ToWIF(key.Key, "simnet")
	address, _ := GetAddressByWIF(prikWIF)
	return address
}

// 根据PublicKey获取对应的地址
func GetAddressByPubKey(key *PublicKey) string {
	// 生成地址
	pubKey, err := secp256k1.ParsePubKey(key.Key)
	if err != nil {
		log.Fatalf("get address error %s", err)
	}
	addr, err := btcutil.NewAddressPubKey(pubKey.SerializeUncompressed(), &chaincfg.SimNetParams)
	return addr.EncodeAddress()
}
