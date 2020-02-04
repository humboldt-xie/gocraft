package world

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	dbpath = flag.String("db", "gocraft.db", "db file name")
)

var (
	blockBucket  = []byte("block")
	chunkBucket  = []byte("chunk")
	cameraBucket = []byte("camera")

	store Store
)

func InitBoltStore() error {
	var path string
	if *dbpath != "" {
		path = *dbpath
	}
	if *serverAddr != "" {
		path = fmt.Sprintf("cache_%s.db", *serverAddr)
	}
	if path == "" {
		return errors.New("empty db path")
	}
	var err error
	store, err = NewBoltStore(path)
	return err
}

func CloseStore() {
	store.Close()
}

type Store interface {
	UpdateBlock(id Vec3, w *Block) error
	//UpdatePlayerState(state Position) error
	UpdatePlayer(p *Player) error
	GetPlayer() *Player
	RangeBlocks(id Vec3, f func(bid Vec3, w *Block)) error
	UpdateChunkVersion(id Vec3, version string) error
	GetChunkVersion(id Vec3) string
	Close()
}

type BoltStore struct {
	db *bolt.DB
}

func NewBoltStore(p string) (Store, error) {
	db, err := bolt.Open(p, 0666, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(blockBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(chunkBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(cameraBucket)
		return err
	})
	if err != nil {
		return nil, err
	}
	db.NoSync = true
	return &BoltStore{
		db: db,
	}, nil
}

func (s *BoltStore) UpdateBlock(id Vec3, w *Block) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		log.Printf("put %v -> %d", id, w)
		bkt := tx.Bucket(blockBucket)
		cid := id.Chunkid()
		key := encodeBlockDbKey(cid, id)
		value := encodeBlockDbValue(w)
		return bkt.Put(key, value)
	})
}

func (s *BoltStore) UpdatePlayer(p *Player) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(cameraBucket)
		b, err := json.Marshal(p)
		if err != nil {
			return err
		}
		bkt.Put(cameraBucket, b)
		return nil
	})
}

func (s *BoltStore) GetPlayer() (player *Player) {
	player = NewPlayer(mgl32.Vec3{0, 16, 0}, nil, nil)
	//var state Position
	//state.Vec3[1] = 16
	s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(cameraBucket)
		value := bkt.Get(cameraBucket)
		if value == nil {
			return nil
		}
		err := json.Unmarshal(value, player)
		if err != nil {
			return err
		}
		return nil
	})
	return player
}

func (s *BoltStore) RangeBlocks(id Vec3, f func(bid Vec3, w *Block)) error {
	return s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blockBucket)
		startkey := encodeBlockDbKey(id, Vec3{0, 0, 0})
		iter := bkt.Cursor()
		for k, v := iter.Seek(startkey); k != nil; k, v = iter.Next() {
			cid, bid := decodeBlockDbKey(k)
			if cid != id {
				break
			}
			w := decodeBlockDbValue(v)
			if w != nil {
				f(bid, w)
			}
		}
		return nil
	})
}

func (s *BoltStore) UpdateChunkVersion(id Vec3, version string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(chunkBucket)
		key := encodeVec3(id)
		return bkt.Put(key, []byte(version))
	})
}

func (s *BoltStore) GetChunkVersion(id Vec3) string {
	var version string
	s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(chunkBucket)
		key := encodeVec3(id)
		v := bkt.Get(key)
		if v != nil {
			version = string(v)
		}
		return nil
	})
	return version
}

func (s *BoltStore) Close() {
	s.db.Sync()
	s.db.Close()
}

func encodeVec3(v Vec3) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, [...]int32{int32(v.X), int32(v.Y), int32(v.Z)})
	return buf.Bytes()
}

func encodeBlockDbKey(cid, bid Vec3) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, [...]int32{int32(cid.X), int32(cid.Z)})
	binary.Write(buf, binary.LittleEndian, [...]int32{int32(bid.X), int32(bid.Y), int32(bid.Z)})
	return buf.Bytes()
}

func decodeBlockDbKey(b []byte) (Vec3, Vec3) {
	if len(b) != 4*5 {
		log.Panicf("bad db key length:%d", len(b))
	}
	buf := bytes.NewBuffer(b)
	var arr [5]int32
	binary.Read(buf, binary.LittleEndian, &arr)

	cid := Vec3{int(arr[0]), 0, int(arr[1])}
	bid := Vec3{int(arr[2]), int(arr[3]), int(arr[4])}
	if bid.Chunkid() != cid {
		log.Panicf("bad db key: cid:%v, bid:%v", cid, bid)
	}
	return cid, bid
}

func encodeBlockDbValue(w *Block) []byte {
	value, _ := json.Marshal(w)
	return value
}

func decodeBlockDbValue(b []byte) (r *Block) {
	json.Unmarshal(b, &r)
	return r
}
