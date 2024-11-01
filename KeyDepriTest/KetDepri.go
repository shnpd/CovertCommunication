package main

import (
	"covertCommunication/KeyDerivation"
	"fmt"
)

func main() {
	seed, err := KeyDerivation.NewSeed()
	seed = []byte("111")
	if err != nil {
		return
	}
	masterKey, err := KeyDerivation.GenerateMasterKey(seed)
	if err != nil {
		return
	}
	//key := masterKey.Key

	key12, _ := masterKey.ChildPrivateKeyDeprive(2)
	key13, _ := masterKey.ChildPrivateKeyDeprive(3)
	key14, _ := masterKey.ChildPrivateKeyDeprive(4)
	fmt.Println(KeyDerivation.ToWIF(key12.Key, "simnet"))
	fmt.Println(KeyDerivation.ToWIF(key13.Key, "simnet"))
	fmt.Println(KeyDerivation.ToWIF(key14.Key, "simnet"))

	key12Addr, _ := KeyDerivation.GetAddressByWIF("4MWc2dmFrDhZfjsKZydjkjYym9BycC1y3aHvDCrjjsUuQTfxndZ")
	fmt.Println(key12Addr)
}
