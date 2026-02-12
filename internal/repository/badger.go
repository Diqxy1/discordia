package repository

import (
	"discordia/internal/domain"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

const identityKey = "sys_identity_private_key"

type BadgerRepo struct {
	db *badger.DB
}

func NewBadgerRepo(path string) *BadgerRepo {
	opts := badger.DefaultOptions(path).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	return &BadgerRepo{db: db}
}

func (r *BadgerRepo) Save(msg *domain.Message) error {
	return r.db.Update(func(txn *badger.Txn) error {
		payload, _ := json.Marshal(msg)
		key := []byte(fmt.Sprintf("msg_%d", msg.Timestamp.UnixNano()))
		return txn.Set(key, payload)
	})
}

func (r *BadgerRepo) GetAll() ([]domain.Message, error) {
	var msgs []domain.Message
	err := r.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			item.Value(func(v []byte) error {
				var m domain.Message
				json.Unmarshal(v, &m)
				msgs = append(msgs, m)
				return nil
			})
		}
		return nil
	})
	return msgs, err
}

func (r *BadgerRepo) SaveIdentity(privKey []byte) error {
	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(identityKey), privKey)
	})
}

func (r *BadgerRepo) GetIdentity() ([]byte, error) {
	var privKey []byte
	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(identityKey))
		if err != nil {
			return err
		}
		privKey, err = item.ValueCopy(nil)
		return err
	})
	return privKey, err
}
