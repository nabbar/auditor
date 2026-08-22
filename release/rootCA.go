package release

/*
 **************************************
 *** THIS FILE IS AUTO GENERATED !! ***
 **************************************
 */

import (
	tlscas "github.com/nabbar/golib/certificates/ca"
)

var res tlscas.Cert

func init() {
	res = nil

	for _, c := range GetRootCA() {
		if res == nil {
			res, _ = tlscas.Parse(c)
		} else {
			_ = res.AppendString(c)
		}
	}
}

func GetRootCACert() tlscas.Cert {
	return res
}

func GetRootCA() []string {
	return []string{``}
}
