/*
 *    Copyright 2019 Insolar
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package drop

import (
	"context"
	"io"

	"github.com/dgraph-io/badger"
	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/ledger/storage"
	"github.com/insolar/insolar/ledger/storage/jet"
	"github.com/insolar/insolar/ledger/storage/record"
	"github.com/pkg/errors"
)

// Builder is an helper-interface, that helps to build new jetdrops
//go:generate minimock -i github.com/insolar/insolar/ledger/storage/jet/drop.Builder -o ./ -s _mock.go
type Builder interface {
	Append(item Hashable) error
	Size(size uint64)
	PrevHash(prevHash []byte)
	Pulse(pn core.PulseNumber)

	Build() (jet.Drop, error)
}

// Hashable is a base interface for an item, that can be appended to builder
type Hashable interface {
	WriteHashData(w io.Writer) (int, error)
}

type builder struct {
	core.Hasher
	dropSize *uint64
	prevHash []byte
	pn       *core.PulseNumber
}

// NewBuilder creates a new instance of Builder
func NewBuilder(hasher core.Hasher) Builder {
	return &builder{
		Hasher: hasher,
	}
}

// Append appends a new item to builder
func (b *builder) Append(item Hashable) (err error) {
	_, err = item.WriteHashData(b.Hasher)
	return
}

// Size sets a drop's size
func (b *builder) Size(size uint64) {
	b.dropSize = &size
}

// PrevHash sets a drop's prevHash
func (b *builder) PrevHash(prevHash []byte) {
	b.prevHash = prevHash
}

// Pulse sets a drop's pulse
func (b *builder) Pulse(pn core.PulseNumber) {
	b.pn = &pn
}

// Build builds Drop and returns it
func (b *builder) Build() (jet.Drop, error) {
	if b.pn == nil {
		return jet.Drop{}, errors.New("pulseNumber is required")
	}
	if b.dropSize == nil {
		return jet.Drop{}, errors.New("dropSize is required")
	}
	if b.prevHash == nil && *b.pn != core.FirstPulseNumber {
		return jet.Drop{}, errors.New("prevHash is required")
	}

	return jet.Drop{
		Pulse:    *b.pn,
		PrevHash: b.prevHash,
		Hash:     b.Hasher.Sum(nil),
		DropSize: *b.dropSize,
	}, nil
}

// Packer is an wrapper interface around process of building jetdrop
// It's considered that implementation of packer uses Bulder under the hood
//go:generate minimock -i github.com/insolar/insolar/ledger/storage/jet/drop.Packer -o ./ -s _mock.go
type Packer interface {
	Pack(ctx context.Context, jetID core.JetID, pulse core.PulseNumber, prevHash []byte) (jet.Drop, error)
}

// NewPacker creates db-based impl of packer
func NewPacker(hasher core.Hasher, db storage.DBContext) Packer {
	return &packer{
		Builder:   NewBuilder(hasher),
		DBContext: db,
	}
}

type packer struct {
	Builder
	storage.DBContext
}

// Pack creates new Drop through interactions with db and Builder
func (p *packer) Pack(ctx context.Context, jetID core.JetID, pulse core.PulseNumber, prevHash []byte) (jet.Drop, error) {
	p.DBContext.WaitingFlight()
	_, jetPrefix := jetID.Jet()

	var dropSize uint64
	recordPrefix := storage.IDRecordPrefixKey(jetPrefix, pulse)

	err := p.GetBadgerDB().View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(recordPrefix); it.ValidForPrefix(recordPrefix); it.Next() {
			val, err := it.Item().ValueCopy(nil)
			if err != nil {
				return err
			}

			err = p.Append(record.DeserializeRecord(val))
			if err != nil {
				return err
			}
			dropSize += uint64(len(val))
		}
		return nil
	})
	if err != nil {
		return jet.Drop{}, err
	}

	p.Pulse(pulse)
	p.PrevHash(prevHash)
	p.Size(dropSize)

	return p.Build()
}
