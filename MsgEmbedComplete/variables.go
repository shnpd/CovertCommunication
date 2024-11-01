package main

import (
	"covertCommunication/KeyDerivation"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/rpcclient"
)

var (
	pkroot        *KeyDerivation.PublicKey      //根公钥
	skroot        *KeyDerivation.PrivateKey     //根私钥
	initSeed      []byte                        //初始种子
	prikSet       [][]*KeyDerivation.PrivateKey //私钥集合,每次消息传递需要一个私钥数组
	pubkSet       [][]*KeyDerivation.PublicKey  //公钥集合
	addressSet    [][]string                    //地址集
	client        *rpcclient.Client             //客户端
	commId        = 0                           //通信序号
	kleak         string                        //泄露消息
	Covertmsg     = "Hello! This is Shihaonan, and i will send a message based blockchain, best wishes!"
	bankPriks     []*KeyDerivation.PrivateKey
	bankRoot      *KeyDerivation.PrivateKey
	bankId        = 0 //下一个银行地址序号
	UTXObyAddress = make(map[string][]btcjson.ListUnspentResult)

	layer1Prik []*KeyDerivation.PrivateKey
)
