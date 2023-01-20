package terminusx

import (
	"context"
	"io"
	"sync/atomic"

	"codeberg.org/gruf/go-bytes"
	"codeberg.org/gruf/go-iotools"

  "codeberg.org/gruf/go-store/v2/storage"
)

// TerminusxStorage is a storage implementation that simply stores key-value
// pairs in a Go map in-memory. The map is protected by a mutex.
type TerminusxStorage struct {
	ow bool // overwrites
	fs *KvDocType
	st uint32
}

// OpenTerminusx opens a new TerminusxStorage instance with internal map starting size.
func OpenTerminusx(fs *KvDocType, overwrites bool) *TerminusxStorage {
	return &TerminusxStorage{
		fs: fs,
		ow: overwrites,
	}
}

// Clean implements Storage.Clean().
func (st *TerminusxStorage) Clean(ctx context.Context) error {
	// Check store open
	if st.closed() {
		return storage.ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}

// ReadBytes implements Storage.ReadBytes().
func (st *TerminusxStorage) ReadBytes(ctx context.Context, key string) ([]byte, error) {
	// Check store open
	if st.closed() {
		return nil, storage.ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Check for key in store
	b, err := st.fs.GetB64(key)
	if err != nil {
		return nil, storage.ErrNotFound
	}

	// Create return copy
	return copyb(b), nil
}

// ReadStream implements Storage.ReadStream().
func (st *TerminusxStorage) ReadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	// Check store open
	if st.closed() {
		return nil, storage.ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Check for key in store
	b, err := st.fs.GetB64(key)
	if err != nil {
		return nil, storage.ErrNotFound
	}

	// Create io.ReadCloser from 'b' copy
	r := bytes.NewReader(copyb(b))
	return iotools.NopReadCloser(r), nil
}

// WriteBytes implements Storage.WriteBytes().
func (st *TerminusxStorage) WriteBytes(ctx context.Context, key string, b []byte) (int, error) {
	// Check store open
	if st.closed() {
		return 0, storage.ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	// Check for key that already exists
	if _, err := st.fs.GetB64(key); err == nil && !st.ow {
		return 0, storage.ErrAlreadyExists
	}

	// Write key copy to store
	st.fs.SetB64(key, copyb(b))
	return len(b), nil
}

// WriteStream implements Storage.WriteStream().
func (st *TerminusxStorage) WriteStream(ctx context.Context, key string, r io.Reader) (int64, error) {
	// Check store open
	if st.closed() {
		return 0, storage.ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	// Check for key that already exists
	if _, err := st.fs.GetB64(key); err == nil && !st.ow {
		return 0, storage.ErrAlreadyExists
	}

	// Read all from reader
	b, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}

	// Write key to store
	st.fs.SetB64(key, b)
	return int64(len(b)), nil
}

// Stat implements Storage.Stat().
func (st *TerminusxStorage) Stat(ctx context.Context, key string) (bool, error) {
	// Check store open
	if st.closed() {
		return false, storage.ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return false, err
	}

	// Check for key in store
	_, err := st.fs.GetB64(key)
	return err == nil, err
}

// Remove implements Storage.Remove().
func (st *TerminusxStorage) Remove(ctx context.Context, key string) error {
	// Check store open
	if st.closed() {
		return storage.ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return err
	}

	// Attempt to delete key
	_, err := st.fs.DeleteDocument(key)
	if err != nil {
		return storage.ErrNotFound
	}

	return nil
}

// WalkKeys implements Storage.WalkKeys().
func (st *TerminusxStorage) WalkKeys(ctx context.Context, opts storage.WalkKeysOptions) error {
	// Check store open
	if st.closed() {
		return storage.ErrClosed
	}

	// Check context still valid
	if err := ctx.Err(); err != nil {
		return err
	}
  
	var err error
  /*
	// Nil check func
	_ = opts.WalkFn

	// Pass each key in map to walk function
	st.fs.Range(func(key string, val []byte) bool {
		err = opts.WalkFn(ctx, storage.Entry{
			Key:  key,
			Size: int64(len(val)),
		})
		return (err == nil)
	})
  */
	return err
}

// Close implements Storage.Close().
func (st *TerminusxStorage) Close() error {
	atomic.StoreUint32(&st.st, 1)
	return nil
}

// closed returns whether TerminusxStorage is closed.
func (st *TerminusxStorage) closed() bool {
	return (atomic.LoadUint32(&st.st) == 1)
}

// copyb returns a copy of byte-slice b.
func copyb(b []byte) []byte {
	if b == nil {
		return nil
	}
	p := make([]byte, len(b))
	_ = copy(p, b)
	return p
}
