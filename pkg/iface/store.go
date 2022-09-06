package iface

type Storer interface {
	Get(key string) []byte
	Set(key string, val []byte)
	Delete(key string)
	Close() error
}
