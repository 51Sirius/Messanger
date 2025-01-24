package testmessage

import (
	"github.com/nspcc-dev/neo-go/pkg/interop/native/management"
	"github.com/nspcc-dev/neo-go/pkg/interop/runtime"
	"github.com/nspcc-dev/neo-go/pkg/interop/storage"
)

func _deploy(data interface{}, isUpdate bool) {
	if isUpdate {
		runtime.Log("contract updating")
	}
}

func SendMessage(key string, messageHash string) bool {
	ctx := storage.GetContext()

	m := storage.Get(ctx, key)

	if m != nil {
		runtime.Log("Message is already exist")
		return false
	}

	storage.Put(ctx, key, messageHash)

	return true
}

func FindAndCheckMessage(key string, messageHash string) bool {
	ctx := storage.GetReadOnlyContext()
	m := storage.Get(ctx, key)

	if m == nil {
		runtime.Log("Message does not exist")
		return false
	}

	return m == messageHash
}

func UpdateMessage(key string, messageHash string) {
	ctx := storage.GetContext()
	storage.Put(ctx, key, messageHash)
}

func DeleteMessage(key string) {
	ctx := storage.GetContext()

	if storage.Get(ctx, key) == nil {
		panic("Message does not exist")
	}

	storage.Delete(ctx, key)
	runtime.Log("Message cleared")
}
func Update(scrypt []byte, manifest []byte, data any) {
	management.UpdateWithData(scrypt, manifest, data)
	runtime.Log("message contract updated")
}
