package release

/*
 **************************************
 *** THIS FILE IS AUTO GENERATED !! ***
 **************************************
 */


import (
	libenc "github.com/nabbar/golib/encoding"
	encaes "github.com/nabbar/golib/encoding/aes"
)

const cryptKey = "b69085678c211a4ca8ec988e084b87afb575136da08a4b6b5bd7d13141f536bd"
const cryptNonce = "36608b88613808c71fcd3a6d"

var crp libenc.Coder

func init() {
	var (
		err error
		key [32]byte
		non [12]byte
	)

	if key, err = encaes.GetHexKey(cryptKey); err != nil {
		panic(err)
	}

	if non, err = encaes.GetHexNonce(cryptNonce); err != nil {
		panic(err)
	}

	if crp, err = encaes.New(key, non); err != nil {
		panic(err)
	}
}

func GetCrypt() libenc.Coder {
	return crp
}

