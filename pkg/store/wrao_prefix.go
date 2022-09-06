package store

import (
	"github.com/chyroc/dropbox-to-google-photos/pkg/iface"
)

type wrapPrefix struct {
	prefix string
	store  iface.Storer
}

func WrapPrefixStore(prefix string, store iface.Storer) iface.Storer {
	return &wrapPrefix{
		prefix: prefix,
		store:  store,
	}
}

func (r *wrapPrefix) Get(key string) []byte {
	return r.store.Get(r.prefix + key)
}

func (r *wrapPrefix) Set(key string, val []byte) {
	r.store.Set(r.prefix+key, val)
}

func (r *wrapPrefix) Delete(key string) {
	r.store.Delete(r.prefix + key)
}

func (r *wrapPrefix) Close() error {
	return r.store.Close()
}
