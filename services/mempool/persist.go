package mempool

import (
	"bytes"
	"fmt"
	"github.com/Qitmeer/qitmeer/core/dbnamespace"
	"github.com/Qitmeer/qitmeer/core/types"
	l "github.com/Qitmeer/qitmeer/log"
	"github.com/schollz/progressbar/v3"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	MempoolFileName = "mempool"
	MempoolVersion  = 0x01
)

func (mp *TxPool) Save() (int, error) {
	txds := mp.TxDescs()
	if len(txds) <= 0 {
		log.Info("There are no transactions to save in mempool.")
		return 0, nil
	}
	outFilePath := filepath.Join(mp.cfg.DataDir, MempoolFileName)
	outFile, err := os.OpenFile(outFilePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	if err != nil {
		return 0, err
	}
	defer func() {
		outFile.Close()
	}()
	//
	_, err = outFile.Write([]byte{byte(MempoolVersion)})
	if err != nil {
		return 0, err
	}
	var serializedBytes [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedBytes[:], uint32(len(txds)))
	_, err = outFile.Write(serializedBytes[:])
	if err != nil {
		return 0, err
	}

	for _, txd := range txds {
		mtd := MempoolTxData{tx: txd.Tx.Tx, time: &txd.Added}
		err := mtd.Encode(outFile)
		if err != nil {
			return 0, err
		}
	}
	//
	return len(txds), nil
}

func (mp *TxPool) Load() error {
	if !mp.cfg.Persist {
		return nil
	}
	outFilePath := filepath.Join(mp.cfg.DataDir, MempoolFileName)
	_, err := os.Stat(outFilePath)
	if err != nil {
		if !os.IsExist(err) {
			log.Debug(err.Error())
			return nil
		}
	}

	bs, err := ioutil.ReadFile(outFilePath)
	if err != nil {
		return err
	}

	offset := 1
	version := bs[0]
	if version != MempoolVersion {
		return fmt.Errorf("The version(%d) of the file does not match %d\n", version, MempoolVersion)
	}
	txNum := dbnamespace.ByteOrder.Uint32(bs[offset : offset+4])
	offset += 4
	var bar *progressbar.ProgressBar
	logLvl := l.Glogger().GetVerbosity()
	if !mp.cfg.NoMempoolBar {
		bar = progressbar.Default(int64(txNum), "Mempool load:")
		l.Glogger().Verbosity(l.LvlCrit)
	}
	add := 0
	for i := uint32(0); i < txNum; i++ {
		if bar != nil {
			bar.Add(1)
		}
		mtd := &MempoolTxData{}
		off, err := mtd.Decode(bs[offset:])
		if err != nil {
			return fmt.Errorf("Mempool load error: tx=%d bytes=%d/%d\n", i, off, offset)
		}
		offset += off
		//
		if time.Since(*mtd.time) > mp.cfg.Expiry {
			log.Info(fmt.Sprintf("Mempool add %s from %s is expiry(%s)", mtd.tx.TxHash().String(), outFilePath, mtd.time.String()))
			continue
		}
		//
		allowOrphans := mp.cfg.Policy.MaxOrphanTxs > 0
		acceptedTxs, err := mp.ProcessTransaction(types.NewTx(mtd.tx), allowOrphans, true, true)
		if err != nil {
			return fmt.Errorf("Failed to process transaction %v: %v\n", mtd.tx.TxHash().String(), err.Error())
		}
		for _, tx := range acceptedTxs {
			log.Debug(fmt.Sprintf("Mempool add %s from %s", tx.Tx.Hash().String(), outFilePath))
		}
		add += len(acceptedTxs)
	}
	l.Glogger().Verbosity(logLvl)
	log.Info(fmt.Sprintf("Mempool load:%d/%d", add, txNum))
	return os.Remove(outFilePath)
}

func (mp *TxPool) IsPersist() bool {
	return mp.cfg.Persist
}

func (mp *TxPool) Perisit() (int, error) {
	mp.cfg.Persist = true
	return mp.Save()
}

type MempoolTxData struct {
	tx   *types.Transaction
	time *time.Time
}

// encode
func (mtd *MempoolTxData) Encode(w io.Writer) error {
	bytes, err := mtd.tx.Serialize()
	if err != nil {
		return err
	}
	var serializedBytes [4]byte
	dbnamespace.ByteOrder.PutUint32(serializedBytes[:], uint32(len(bytes)))
	_, err = w.Write(serializedBytes[:])
	if err != nil {
		return err
	}
	l, err := w.Write(bytes)
	if l != len(bytes) {
		return fmt.Errorf("mem pool persist:%s", mtd.tx.TxHash().String())
	}

	var timeBytes [8]byte
	dbnamespace.ByteOrder.PutUint64(timeBytes[:], uint64(mtd.time.Unix()))
	_, err = w.Write(timeBytes[:])
	return err
}

// decode
func (mtd *MempoolTxData) Decode(bs []byte) (int, error) {
	offset := 0
	txSize := int(dbnamespace.ByteOrder.Uint32(bs[offset : offset+4]))
	offset += 4

	txBytes := bs[offset : offset+txSize]
	offset += txSize

	mtd.tx = &types.Transaction{}
	err := mtd.tx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		return offset, err
	}
	ti := int64(dbnamespace.ByteOrder.Uint64(bs[offset : offset+8]))
	offset += 8

	t := time.Unix(ti, 0)
	mtd.time = &t
	return offset, nil
}
