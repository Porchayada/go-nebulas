// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package core

import (
	"testing"

	"time"

	"github.com/nebulasio/go-nebulas/crypto"
	"github.com/nebulasio/go-nebulas/crypto/keystore"
	"github.com/nebulasio/go-nebulas/crypto/keystore/secp256k1"
	"github.com/nebulasio/go-nebulas/util"
	"github.com/stretchr/testify/assert"
)

func TestTransactionPool(t *testing.T) {
	ks := keystore.DefaultKS
	priv1, _ := secp256k1.GeneratePrivateKey()
	pubdata1, _ := priv1.PublicKey().Encoded()
	from, _ := NewAddressFromPublicKey(pubdata1)
	ks.SetKey(from.ToHex(), priv1, []byte("passphrase"))
	ks.Unlock(from.ToHex(), []byte("passphrase"), time.Second*60*60*24*365)
	key1, _ := ks.GetUnlocked(from.ToHex())
	signature1, _ := crypto.NewSignature(keystore.SECP256K1)
	signature1.InitSign(key1.(keystore.PrivateKey))

	priv2, _ := secp256k1.GeneratePrivateKey()
	pubdata2, _ := priv2.PublicKey().Encoded()
	other, _ := NewAddressFromPublicKey(pubdata2)
	ks.SetKey(other.ToHex(), priv2, []byte("passphrase"))
	ks.Unlock(other.ToHex(), []byte("passphrase"), time.Second*60*60*24*365)
	key2, _ := ks.GetUnlocked(other.ToHex())
	signature2, _ := crypto.NewSignature(keystore.SECP256K1)
	signature2.InitSign(key2.(keystore.PrivateKey))

	txs := []*Transaction{
		NewTransaction(1, from, &Address{[]byte("to")}, util.NewUint128(), 10, []byte("datadata")),
		NewTransaction(1, other, &Address{[]byte("to")}, util.NewUint128(), 1, []byte("datadata")),
		NewTransaction(1, from, &Address{[]byte("to")}, util.NewUint128(), 1, []byte("da")),

		NewTransaction(1, from, &Address{[]byte("to")}, util.NewUint128(), 2, []byte("da")),
		NewTransaction(0, from, &Address{[]byte("to")}, util.NewUint128(), 0, []byte("da")),

		NewTransaction(1, other, &Address{[]byte("to")}, util.NewUint128(), 1, []byte("data")),
		NewTransaction(1, from, &Address{[]byte("to")}, util.NewUint128(), 1, []byte("datadata")),
	}

	txPool := NewTransactionPool(3)
	bc := NewBlockChain(1)
	txPool.setBlockChain(bc)
	txs[0].Sign(signature1)
	assert.Nil(t, txPool.Push(txs[0]))
	// put dup tx, should fail
	assert.NotNil(t, txPool.Push(txs[0]))
	txs[1].Sign(signature2)
	assert.Nil(t, txPool.Push(txs[1]))
	txs[2].Sign(signature1)
	assert.Nil(t, txPool.Push(txs[2]))
	// put not signed tx, should fail
	assert.NotNil(t, txPool.Push(txs[3]))
	// put tx with different chainID, should fail
	txs[4].Sign(signature1)
	assert.NotNil(t, txPool.Push(txs[4]))
	// put one new, replace txs[1]
	assert.Equal(t, len(txPool.all), 3)
	assert.Equal(t, txPool.cache.Len(), 3)
	txs[6].Sign(signature1)
	assert.Nil(t, txPool.Push(txs[6]))
	assert.Equal(t, txPool.cache.Len(), 3)
	assert.Equal(t, len(txPool.all), 3)
	// get from: other, nonce: 1, data: "da"
	tx1 := txPool.Pop()
	assert.Equal(t, txs[2].from.address, tx1.from.address)
	assert.Equal(t, txs[2].nonce, tx1.nonce)
	assert.Equal(t, txs[2].data, tx1.data)
	// put one new
	assert.Equal(t, len(txPool.all), 2)
	assert.Equal(t, txPool.cache.Len(), 2)
	txs[5].Sign(signature2)
	assert.Nil(t, txPool.Push(txs[5]))
	assert.Equal(t, len(txPool.all), 3)
	assert.Equal(t, txPool.cache.Len(), 3)
	// get 2 txs, txs[5], txs[0]
	tx21 := txPool.Pop()
	tx22 := txPool.Pop()
	assert.Equal(t, txs[5].from.address, tx21.from.address)
	assert.Equal(t, txs[5].Nonce(), tx21.Nonce())
	assert.Equal(t, txs[5].data, tx21.data)
	assert.Equal(t, txs[6].from.address, tx22.from.address)
	assert.Equal(t, txs[6].Nonce(), tx22.Nonce())
	assert.Equal(t, txs[6].data, tx22.data)
	assert.Equal(t, txPool.Empty(), false)
	txPool.Pop()
	assert.Equal(t, txPool.Empty(), true)
	assert.Nil(t, txPool.Pop())
}