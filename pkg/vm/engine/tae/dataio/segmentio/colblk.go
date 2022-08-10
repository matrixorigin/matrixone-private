// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package segmentio

import (
	"fmt"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/types"

	"github.com/matrixorigin/matrixone/pkg/compress"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/iface/file"
)

type columnBlock struct {
	common.RefHelper
	block   *blockFile
	ts      types.TS
	indexes []*indexFile
	updates *updatesFile
	data    *dataFile
	col     int
}

func newColumnBlock(block *blockFile, indexCnt int, col int) *columnBlock {
	cb := &columnBlock{
		block:   block,
		indexes: make([]*indexFile, indexCnt),
		col:     col,
	}
	for i := range cb.indexes {
		cb.indexes[i] = newIndex(cb)
		cb.indexes[i].dataFile.file = append(cb.indexes[i].dataFile.file, cb.block.seg.GetSegmentFile().NewBlockFile(
			fmt.Sprintf("%d_%d_%d.idx", cb.col, cb.block.id, i)))
		cb.indexes[i].dataFile.file[0].snode.algo = compress.None
		cb.indexes[i].dataFile.file[0].SetIdxs(uint32(len(cb.indexes)))
		cb.indexes[i].dataFile.file[0].SetCols(uint32(col))
	}
	cb.updates = newUpdates(cb)
	cb.data = newData(cb)
	cb.OnZeroCB = cb.close
	cb.Ref()
	return cb
}

func (cb *columnBlock) AddIndex(idx int) {
	idxCnt := len(cb.indexes)
	if idx > idxCnt {
		for i := idxCnt; i < idx; i++ {
			cb.indexes = append(cb.indexes, newIndex(cb))
		}
	}
}

func (cb *columnBlock) WriteTS(ts types.TS) (err error) {
	cb.ts = ts
	cb.data.SetFile(
		//cb.block.seg.GetSegmentFile().NewBlockFile(fmt.Sprintf("%d_%d_%d.blk", cb.col, cb.block.id, ts)),
		cb.block.seg.GetSegmentFile().NewBlockFile(fmt.Sprintf("%d_%d_%s.blk", cb.col,
			cb.block.id, ts.ToString())),
		uint32(len(cb.block.columns)),
		uint32(len(cb.indexes)))
	cb.updates.SetFile(
		//cb.block.seg.GetSegmentFile().NewBlockFile(fmt.Sprintf("%d_%d_%d.update", cb.col, cb.block.id, ts)),
		cb.block.seg.GetSegmentFile().NewBlockFile(fmt.Sprintf("%d_%d_%s.update", cb.col,
			cb.block.id, ts.ToString())),
		uint32(len(cb.block.columns)),
		uint32(len(cb.indexes)))
	return
}

func (cb *columnBlock) WriteData(buf []byte) (err error) {
	_, err = cb.data.Write(buf)
	return
}

func (cb *columnBlock) WriteUpdates(buf []byte) (err error) {
	_, err = cb.updates.Write(buf)
	return
}

func (cb *columnBlock) WriteIndex(idx int, buf []byte) (err error) {
	if idx >= len(cb.indexes) {
		err = file.ErrInvalidParam
		return
	}
	vfile := cb.indexes[idx]
	_, err = vfile.Write(buf)
	return
}

func (cb *columnBlock) ReadTS() types.TS { return cb.ts }

func (cb *columnBlock) ReadData(buf []byte) (err error) {
	_, err = cb.data.Read(buf)
	return
}

func (cb *columnBlock) ReadUpdates(buf []byte) (err error) {
	_, err = cb.updates.Read(buf)
	return
}

func (cb *columnBlock) ReadIndex(idx int, buf []byte) (err error) {
	if idx >= len(cb.indexes) {
		err = file.ErrInvalidParam
		return
	}
	vfile := cb.indexes[idx]
	_, err = vfile.Read(buf)
	return
}

func (cb *columnBlock) GetDataFileStat() (stat common.FileInfo) {
	return cb.data.stat
}

func (cb *columnBlock) OpenIndexFile(idx int) (vfile common.IRWFile, err error) {
	if idx >= len(cb.indexes) {
		err = file.ErrInvalidParam
		return
	}
	vfile = cb.indexes[idx]
	vfile.Ref()
	return
}

func (cb *columnBlock) OpenUpdateFile() (vfile common.IRWFile, err error) {
	cb.updates.Ref()
	vfile = cb.updates
	return
}

func (cb *columnBlock) OpenDataFile() (vfile common.IRWFile, err error) {
	cb.data.Ref()
	vfile = cb.data
	return
}

func (cb *columnBlock) Close() error {
	cb.Unref()
	// cb.data.Unref()
	// cb.updates.Unref()
	// for _, index := range cb.indexes {
	// 	index.Unref()
	// }
	return nil
}

func (cb *columnBlock) close() {
	cb.Destroy()
}

func (cb *columnBlock) Destroy() {
	logutil.Infof("Destroying Block %d Col @ TS %d", cb.block.id, cb.ts)
	cb.data.Destroy()
	cb.data = nil
	for _, index := range cb.indexes {
		index.Destroy()
	}
	cb.indexes = nil
	cb.updates.Destroy()
	cb.updates = nil
}
