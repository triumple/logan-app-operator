package hash

import (
	"github.com/davecgh/go-spew/spew"
	"hash"
)

// DeepHashObject writes specified object to hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
// from https://github.com/kubernetes/kubernetes/blob/28e800245e910b65b56548f36172ce525a554dc8/pkg/util/hash/hash.go
func DeepHashObject(hasher hash.Hash, objectToWrite interface{}) {
	hasher.Reset()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", objectToWrite)
}
