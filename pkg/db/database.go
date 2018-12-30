package db

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	writePauseWarningThrottler = 1 * time.Minute
)

var OpenFileLimit = 64

type LevelDB struct {
	fn string      // filename for reporting
	db *leveldb.DB // LevelDB instance

	quitLock sync.Mutex      // Mutex protecting the quit channel access
	quitChan chan chan error // Quit channel to stop the metrics collection before closing the database
}

// NewLDBDatabase returns a LevelDB wrapped object.
func NewLDBDatabase(file string, cache int, handles int) (*LevelDB, error) {
	// Ensure we have some minimal caching and file guarantees
	if cache < 16 {
		cache = 16
	}
	if handles < 16 {
		handles = 16
	}
	log.Infof("dbfile:%s allocated cache:%d and file handles:%d", file, cache, handles)

	// Open the db and recover any potential corruptions
	db, err := leveldb.OpenFile(file, &opt.Options{
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB, // Two of these are used internally
		Filter:                 filter.NewBloomFilter(10),
	})
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(file, nil)
	}
	// (Re)check for errors and abort if opening of the db failed
	if err != nil {
		return nil, err
	}
	return &LevelDB{
		fn: file,
		db: db,
	}, nil
}

// Path returns the path to the database directory.
func (db *LevelDB) Path() string {
	return db.fn
}

// Put puts the given key / value to the queue
func (db *LevelDB) Put(key []byte, value []byte) error {
	return db.db.Put(key, value, nil)
}

func (db *LevelDB) Has(key []byte) (bool, error) {
	return db.db.Has(key, nil)
}

// Get returns the given key if it's present.
func (db *LevelDB) Get(key []byte) ([]byte, error) {
	dat, err := db.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

// Delete deletes the key from the queue and database
func (db *LevelDB) Delete(key []byte) error {
	return db.db.Delete(key, nil)
}

func (db *LevelDB) NewIterator() iterator.Iterator {
	return db.db.NewIterator(nil, nil)
}

// NewIteratorWithPrefix returns a iterator to iterate over subset of database content with a particular prefix.
func (db *LevelDB) NewIteratorWithPrefix(prefix []byte) iterator.Iterator {
	return db.db.NewIterator(util.BytesPrefix(prefix), nil)
}

func (db *LevelDB) Close() {
	// Stop the metrics collection to avoid internal database races
	db.quitLock.Lock()
	defer db.quitLock.Unlock()

	if db.quitChan != nil {
		errc := make(chan error)
		db.quitChan <- errc
		if err := <-errc; err != nil {
			log.Errorf("Metrics collection failed err: %v", err)
		}
		db.quitChan = nil
	}
	err := db.db.Close()
	if err == nil {
		log.Info("Database closed")
	} else {
		log.Errorf("Failed to close database err: %v", err)
	}
}

func (db *LevelDB) LDB() *leveldb.DB {
	return db.db
}

type LevelDBBatch struct {
	db   *leveldb.DB
	b    *leveldb.Batch
	size int
}

func (db *LevelDB) NewBatch() *LevelDBBatch {
	return &LevelDBBatch{db: db.db, b: new(leveldb.Batch)}
}

func (b *LevelDBBatch) Put(key, value []byte) error {
	b.b.Put(key, value)
	b.size += len(value)
	return nil
}

func (b *LevelDBBatch) Delete(key []byte) error {
	b.b.Delete(key)
	b.size += 1
	return nil
}

func (b *LevelDBBatch) Write() error {
	return b.db.Write(b.b, nil)
}

func (b *LevelDBBatch) ValueSize() int {
	return b.size
}

func (b *LevelDBBatch) Reset() {
	b.b.Reset()
	b.size = 0
}
