package tmsp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	sm "github.com/eris-ltd/eris-db/state"
	types "github.com/eris-ltd/eris-db/txs"

	"github.com/tendermint/go-events"
	client "github.com/tendermint/go-rpc/client"
	"github.com/tendermint/go-wire"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	tmsp "github.com/tendermint/tmsp/types"
)

//--------------------------------------------------------------------------------
// ErisDBApp holds the current state, runs transactions, computes hashes.
// Typically two connections are opened by the tendermint core:
// one for mempool, one for consensus.

type ErisDBApp struct {
	mtx sync.Mutex

	state      *sm.State
	cache      *sm.BlockCache
	checkCache *sm.BlockCache // for CheckTx (eg. so we get nonces right)

	evc  *events.EventCache
	evsw *events.EventSwitch

	// client to the tendermint core rpc
	client *client.ClientURI
	host   string // tendermint core endpoint
}

func (app *ErisDBApp) GetState() *sm.State {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	return app.state.Copy()
}

func (app *ErisDBApp) GetCheckCache() *sm.BlockCache {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	return app.checkCache
}

func (app *ErisDBApp) ResetCheckCache() {
	app.mtx.Lock()
	defer app.mtx.Unlock()
	app.checkCache = sm.NewBlockCache(app.state)
}

func (app *ErisDBApp) SetHostAddress(host string) {
	app.host = host
	app.client = client.NewClientURI(host) //fmt.Sprintf("http://%s", host))
}

// Broadcast a tx to the tendermint core
// NOTE: this assumes we know the address of core
func (app *ErisDBApp) BroadcastTx(tx types.Tx) error {
	var result ctypes.TMResult
	buf := new(bytes.Buffer)
	var n int
	var err error
	wire.WriteBinary(struct{ types.Tx }{tx}, buf, &n, &err)
	if err != nil {
		return err
	}

	params := map[string]interface{}{
		"tx": hex.EncodeToString(buf.Bytes()),
	}
	_, err = app.client.Call("broadcast_tx_sync", params, &result)
	return err
}

func NewErisDBApp(s *sm.State, evsw *events.EventSwitch) *ErisDBApp {
	return &ErisDBApp{
		state:      s,
		cache:      sm.NewBlockCache(s),
		checkCache: sm.NewBlockCache(s),
		evc:        events.NewEventCache(evsw),
		evsw:       evsw,
	}
}

// Implements tmsp.Application
func (app *ErisDBApp) Info() (info string) {
	return "ErisDB"
}

// Implements tmsp.Application
func (app *ErisDBApp) SetOption(key string, value string) (log string) {
	return ""
}

// Implements tmsp.Application
func (app ErisDBApp) AppendTx(txBytes []byte) (res tmsp.Result) {
	// XXX: if we had tx ids we could cache the decoded txs on CheckTx
	var n int
	var err error
	tx := new(types.Tx)
	buf := bytes.NewBuffer(txBytes)
	wire.ReadBinaryPtr(tx, buf, len(txBytes), &n, &err)
	if err != nil {
		return tmsp.NewError(tmsp.CodeType_EncodingError, fmt.Sprintf("Encoding error: %v", err))
	}

	err = sm.ExecTx(app.cache, *tx, true, app.evc)
	if err != nil {
		return tmsp.NewError(tmsp.CodeType_InternalError, fmt.Sprintf("Encoding error: %v", err))
	}
	return tmsp.NewResultOK(nil, "Success")
}

// Implements tmsp.Application
func (app ErisDBApp) CheckTx(txBytes []byte) (res tmsp.Result) {
	log.Info("Check Tx", "tx", txBytes)
	defer func() {
		log.Info("Check Tx", "res", res)
	}()
	var n int
	var err error
	tx := new(types.Tx)
	buf := bytes.NewBuffer(txBytes)
	wire.ReadBinaryPtr(tx, buf, len(txBytes), &n, &err)
	if err != nil {
		return tmsp.NewError(tmsp.CodeType_EncodingError, fmt.Sprintf("Encoding error: %v", err))
	}

	// we need the lock because CheckTx can run concurrently with Commit,
	// and Commit refreshes the checkCache
	app.mtx.Lock()
	defer app.mtx.Unlock()
	err = sm.ExecTx(app.checkCache, *tx, false, nil)
	if err != nil {
		return tmsp.NewError(tmsp.CodeType_InternalError, fmt.Sprintf("Encoding error: %v", err))
	}

	return tmsp.NewResultOK(nil, "Success")
}

// Implements tmsp.Application
// Commit the state (called at end of block)
func (app *ErisDBApp) Commit() (res tmsp.Result) {
	// sync the AppendTx cache
	app.cache.Sync()

	// reset the check cache to the new height
	app.ResetCheckCache()

	// save state to disk
	app.state.Save()

	// flush events to listeners (XXX: note issue with blocking)
	app.evc.Flush()

	return tmsp.NewResultOK(app.state.Hash(), "Success")
}

func (app *ErisDBApp) Query(query []byte) (res tmsp.Result) {
	return tmsp.NewResultOK(nil, "Success")
}