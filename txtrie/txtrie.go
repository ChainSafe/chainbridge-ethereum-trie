// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package txtrie

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/ethereum/go-ethereum/rlp"
	ethtrie "github.com/ethereum/go-ethereum/trie"
)

// TxTries stores all the instances of tries we have on disk
type TxTries struct {
	// TODO: the memory allocated for these is hard to get back, look for better way to have a queue
	txTries map[common.Hash]*ethtrie.Trie
	// txTries      []*ethtrie.Trie
	txRoots      []common.Hash // needed to track insertion order
	triesToStore int
}

var (
	// from https://github.com/ethereum/go-ethereum/blob/bcb308745010675671991522ad2a9e811938d7fb/trie/trie.go#L32
	emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

// NewTxTries creates a new instance of a TxTries object
func NewTxTries(t int) *TxTries {
	txTrie := &TxTries{
		txTries:      make(map[common.Hash]*ethtrie.Trie),
		triesToStore: t,
	}
	return txTrie

}

func (t *TxTries) updateTriesAndRoots(trie *ethtrie.Trie, root common.Hash) error {
	if len(t.txTries) >= t.triesToStore {
		// delete contents of trie from database
		trieToDelete := t.txTries[t.txRoots[0]]
		err := deleteTrie(trieToDelete)
		if err != nil {
			return err
		}
		delete(t.txTries, t.txRoots[0])
		t.txRoots = append(t.txRoots, root)
		t.txTries[root] = trie
		t.txRoots = t.txRoots[1:]

	} else {
		t.txRoots = append(t.txRoots, root)
		t.txTries[root] = trie

	}

	return nil

}

func deleteTrie(trie *ethtrie.Trie) error {
	i := 0

	for {
		// key of transaction
		key, err := rlp.EncodeToBytes(uint(i))
		if err != nil {
			return err
		}

		err = trie.TryDelete(key)
		if err != nil {
			return err
		}

		if trie.Hash() == emptyRoot {
			// eventually we will reach a point where the hash of the root node of the trie is the emptyRoot
			break
		}

		i++
	}

	return nil

}

// AddNewTrie adds a new transaction trie to an existing TxTries object
func (t *TxTries) AddNewTrie(root common.Hash, transactions types.Transactions, db *leveldb.Database) error {

	//if db == nil {
	//	return errors.New("db does not exist")
	//}

	if transactions == nil {
		return errors.New("transactions cannot be nil")
	}

	_, err := t.newTrie(root, db, transactions)

	if err != nil {
		return err
	}

	return nil

}

// AddTrie creates a new instance of a trie object
func (t *TxTries) newTrie(root common.Hash, db *leveldb.Database, transactions types.Transactions) (*ethtrie.Trie, error) {
	// TODO: look into cache values
	// this creates a new trie database with our KVDB as the diskDB for node storage
	trie, err := ethtrie.New(emptyRoot, ethtrie.NewDatabaseWithCache(db, 0, ""))
	if err != nil {
		return nil, err
	}

	err = updateTrie(trie, transactions, root)

	if err != nil {
		return nil, err
	}

	err = t.updateTriesAndRoots(trie, root)

	if err != nil {
		return nil, err
	}

	return trie, nil
}

// updateTrie updates the transaction trie with root transactionRoot with given transactions
// note that this assumes the slice transactions is in the same order they are in the block
func updateTrie(trie *ethtrie.Trie, transactions types.Transactions, transactionRoot common.Hash) error {
	for i, tx := range transactions {

		key, err := rlp.EncodeToBytes(uint(i))
		if err != nil {
			return err
		}

		value, err := rlp.EncodeToBytes(tx)
		if err != nil {
			return err
		}

		trie.Update(key, value)
	}

	// check if the root hash of the trie matches the transactionRoot
	if trie.Hash().Hex() != transactionRoot.Hex() {
		return errors.New("transaction roots don't match")
	}

	return nil
}

// RetrieveEncodedProof retrieves an encoded Proof for a value at key in trie with root root
func (t *TxTries) RetrieveEncodedProof(root common.Hash, key []byte) ([]byte, error) {
	proofDB, err := t.RetrieveProof(root, key)
	if err != nil {
		return nil, err
	}
	return encodeProofDB(root, key, proofDB)
}

// RetrieveProof retrieves a Proof for a value at key in trie with root root
func (t *TxTries) RetrieveProof(root common.Hash, key []byte) (*ProofDatabase, error) {
	trieToRetrieve := t.txTries[root]

	if trieToRetrieve == nil {
		return nil, errors.New("transaction trie for this transaction root does not exist")
	}

	return retrieveProof(trieToRetrieve, root, key)
}

func retrieveProof(trie *ethtrie.Trie, root common.Hash, key []byte) (*ProofDatabase, error) {
	var proof = NewProofDatabase()
	err := trie.Prove(key, 0, proof)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// VerifyProof verifies merkle proof on path key against the provided root
func VerifyProof(root common.Hash, key []byte, proof *ProofDatabase) (bool, error) {
	exists, err := ethtrie.VerifyProof(root, key, proof)

	if err != nil {
		return false, err
	}

	return exists != nil, nil
}
