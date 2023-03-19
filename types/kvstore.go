package types

import (
	"bytes"
	"github.com/dgraph-io/badger"
	tdabcitypes "github.com/tendermint/tendermint/abci/types"
)

type KVStoreApp struct {
	db           *badger.DB
	currentBatch *badger.Txn
}

var _ tdabcitypes.Application = (*KVStoreApp)(nil)

func NewKVStoreApp(db *badger.DB) *KVStoreApp {
	return &KVStoreApp{db: db}
}

func (app *KVStoreApp) SetOption(req tdabcitypes.RequestSetOption) tdabcitypes.ResponseSetOption {
	return tdabcitypes.ResponseSetOption{}
}

func (app *KVStoreApp) Info(req tdabcitypes.RequestInfo) tdabcitypes.ResponseInfo {
	return tdabcitypes.ResponseInfo{}
}

func (app *KVStoreApp) Query(req tdabcitypes.RequestQuery) (resQuery tdabcitypes.ResponseQuery) {
	resQuery.Key = req.Data
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(resQuery.Key)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}

			resQuery.Log = "not exists"
			return nil
		}
		return item.Value(func(val []byte) error {
			resQuery.Log = "exists"
			resQuery.Value = val
			return nil
		})
	})
	if err != nil {
		panic(err)
	}

	return
}

func (app *KVStoreApp) CheckTx(req tdabcitypes.RequestCheckTx) tdabcitypes.ResponseCheckTx {
	code := app.isValid(req.Tx)
	return tdabcitypes.ResponseCheckTx{Code: code, GasWanted: 1}
}

func (app *KVStoreApp) InitChain(req tdabcitypes.RequestInitChain) tdabcitypes.ResponseInitChain {
	return tdabcitypes.ResponseInitChain{}
}

func (app *KVStoreApp) BeginBlock(req tdabcitypes.RequestBeginBlock) tdabcitypes.ResponseBeginBlock {
	app.currentBatch = app.db.NewTransaction(true)
	return tdabcitypes.ResponseBeginBlock{}
}

func (app *KVStoreApp) DeliverTx(req tdabcitypes.RequestDeliverTx) tdabcitypes.ResponseDeliverTx {
	code := app.isValid(req.Tx)
	if code != 0 {
		return tdabcitypes.ResponseDeliverTx{Code: code}
	}

	parts := bytes.Split(req.Tx, []byte("="))
	key, value := parts[0], parts[1]
	err := app.currentBatch.Set(key, value)
	if err != nil {
		panic(err)
	}

	return tdabcitypes.ResponseDeliverTx{Code: code}
}

func (app *KVStoreApp) EndBlock(req tdabcitypes.RequestEndBlock) tdabcitypes.ResponseEndBlock {
	return tdabcitypes.ResponseEndBlock{}
}

func (app *KVStoreApp) Commit() tdabcitypes.ResponseCommit {
	app.currentBatch.Commit()
	return tdabcitypes.ResponseCommit{Data: []byte{}}
}

func (app *KVStoreApp) ListSnapshots(req tdabcitypes.RequestListSnapshots) tdabcitypes.ResponseListSnapshots {
	return tdabcitypes.ResponseListSnapshots{}
}

func (app *KVStoreApp) OfferSnapshot(req tdabcitypes.RequestOfferSnapshot) tdabcitypes.ResponseOfferSnapshot {
	return tdabcitypes.ResponseOfferSnapshot{}
}

func (app *KVStoreApp) LoadSnapshotChunk(req tdabcitypes.RequestLoadSnapshotChunk) tdabcitypes.ResponseLoadSnapshotChunk {
	return tdabcitypes.ResponseLoadSnapshotChunk{}
}

func (app *KVStoreApp) ApplySnapshotChunk(req tdabcitypes.RequestApplySnapshotChunk) tdabcitypes.ResponseApplySnapshotChunk {
	return tdabcitypes.ResponseApplySnapshotChunk{}
}

// isValid checks whether tx is valid.
func (app *KVStoreApp) isValid(tx []byte) (code uint32) {
	// check format
	parts := bytes.Split(tx, []byte("="))
	if len(parts) != 2 {
		return 1
	}

	key, value := parts[0], parts[1]

	// check if the same key=value already exists
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)

		if err != nil {
			if err != badger.ErrKeyNotFound { // 其他错误
				return err
			}
			// key不存在
			return nil
		}

		// key存在
		return item.Value(func(val []byte) error {
			if bytes.Equal(val, value) {
				code = 2
			}
			return nil
		})
	})

	if err != nil {
		panic(err)
	}

	return code
}
