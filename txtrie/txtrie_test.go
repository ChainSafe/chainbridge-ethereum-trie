// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package txtrie

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

var (
	defaultTriesToStore = 3
	emptyHash           = common.HexToHash("")
)

func addTrie(txTries *TxTries, root common.Hash, transactions types.Transactions) error {
	if transactions == nil {
		transactions = types.Transactions{}
	}

	err := txTries.CreateNewTrie(root, transactions)

	if err != nil {
		return err
	}

	return nil

}

func computeEthReferenceTrieHash(transactions types.Transactions) (common.Hash, error) {
	newTrie, err := trie.New(emptyRoot, trie.NewDatabase(nil))
	if err != nil {
		return emptyHash, err
	}

	for i, tx := range transactions {

		key, err := rlp.EncodeToBytes(uint(i))
		if err != nil {
			return emptyHash, err
		}

		value, err := rlp.EncodeToBytes(tx)
		if err != nil {
			return emptyHash, err
		}

		err = newTrie.TryUpdate(key, value)
		if err != nil {
			return emptyHash, err
		}
	}

	return newTrie.Hash(), nil

}

func TestAddEmptyTrie(t *testing.T) {
	txTries := NewTxTries()
	err := addTrie(txTries, emptyRoot, nil)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != emptyRoot {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", emptyRoot, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != emptyRoot {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", emptyRoot, txTries.txTries[txTries.txRoots[0]].Hash())
	}

}

func TestAddEmptyTrieRetrieveProof_Fails(t *testing.T) {
	txTries := NewTxTries()
	err := addTrie(txTries, emptyRoot, nil)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != emptyRoot {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", emptyRoot, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != emptyRoot {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", emptyRoot, txTries.txTries[txTries.txRoots[0]].Hash())
	}
}

func TestAddSingleTrieUpdate(t *testing.T) {
	vals := GetTransactions1()
	expectedRoot, err := computeEthReferenceTrieHash(vals)
	if err != nil {
		t.Fatal(err)
	}

	txTries := NewTxTries()
	err = addTrie(txTries, expectedRoot, vals)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != expectedRoot {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot, txTries.txTries[txTries.txRoots[0]].Hash())
	}
}

func TestAddSingleTrieRetrieveProof(t *testing.T) {
	vals := GetTransactions1()
	expectedRoot, err := computeEthReferenceTrieHash(vals)
	if err != nil {
		t.Fatal(err)
	}

	txTries := NewTxTries()
	err = addTrie(txTries, expectedRoot, vals)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != expectedRoot {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	keyToRetrieve, err := rlp.EncodeToBytes(uint(0))
	if err != nil {
		t.Fatal(err)
	}

	proofDb, err := txTries.RetrieveProof(expectedRoot, keyToRetrieve)
	if err != nil {
		t.Fatal(err)
	}

	exists, err := VerifyProof(expectedRoot, keyToRetrieve, proofDb)
	if err != nil {
		t.Fatal(err)
	}

	if exists != true {
		t.Fatalf("not able to verify retrieved proof!")
	}
}

func TestAddMultipleTries(t *testing.T) {
	txTries := NewTxTries()

	vals1 := GetTransactions1()
	expectedRoot1, err := computeEthReferenceTrieHash(vals1)
	if err != nil {
		t.Fatal(err)
	}

	err = addTrie(txTries, expectedRoot1, vals1)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != expectedRoot1 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot1, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot1 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot1, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	vals2 := GetTransactions2()
	expectedRoot2, err := computeEthReferenceTrieHash(vals2)
	if err != nil {
		t.Fatal(err)
	}

	err = addTrie(txTries, expectedRoot2, vals2)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[1] != expectedRoot2 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot2, txTries.txRoots[1])
	}

	if txTries.txTries[txTries.txRoots[1]].Hash() != expectedRoot2 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot2, txTries.txTries[txTries.txRoots[1]].Hash())
	}

	vals3 := GetTransactions3()
	expectedRoot3, err := computeEthReferenceTrieHash(vals3)
	if err != nil {
		t.Fatal(err)
	}

	err = addTrie(txTries, expectedRoot3, vals3)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[2] != expectedRoot3 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot3, txTries.txRoots[2])
	}

	if txTries.txTries[txTries.txRoots[2]].Hash() != expectedRoot3 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot3, txTries.txTries[txTries.txRoots[2]].Hash())
	}

	err = addTrie(txTries, expectedRoot1, vals1)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[2] != expectedRoot1 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot1, txTries.txRoots[2])
	}

	if txTries.txTries[txTries.txRoots[2]].Hash() != expectedRoot1 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot1, txTries.txTries[txTries.txRoots[2]].Hash())
	}

	if txTries.txRoots[0] != expectedRoot2 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot2, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot2 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot2, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	keyToRetrieve, err := rlp.EncodeToBytes(uint(0))
	if err != nil {
		t.Fatal(err)
	}

	proofDb1, err := txTries.RetrieveProof(expectedRoot1, keyToRetrieve)
	if err != nil {
		t.Fatal(err)
	}

	exists1, err := VerifyProof(expectedRoot1, keyToRetrieve, proofDb1)
	if err != nil {
		t.Fatal(err)
	}

	if exists1 != true {
		t.Fatalf("not able to verify retrieved proof!")
	}

	proofDb2, err := txTries.RetrieveProof(expectedRoot2, keyToRetrieve)
	if err != nil {
		t.Fatal(err)
	}

	exists2, err := VerifyProof(expectedRoot2, keyToRetrieve, proofDb2)
	if err != nil {
		t.Fatal(err)
	}

	if exists2 != true {
		t.Fatalf("not able to verify retrieved proof!")
	}

	proofDb3, err := txTries.RetrieveProof(expectedRoot3, keyToRetrieve)
	if err != nil {
		t.Fatal(err)
	}

	exists3, err := VerifyProof(expectedRoot3, keyToRetrieve, proofDb3)
	if err != nil {
		t.Fatal(err)
	}

	if exists3 != true {
		t.Fatalf("not able to verify retrieved proof!")
	}
}

func TestRetrieveProofDeletedTrie_Fails(t *testing.T) {
	txTries := NewTxTries()
	vals1 := GetTransactions1()
	expectedRoot1, err := computeEthReferenceTrieHash(vals1)
	if err != nil {
		t.Fatal(err)
	}
	err = addTrie(txTries, expectedRoot1, vals1)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != expectedRoot1 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot1, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot1 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot1, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	vals2 := GetTransactions2()
	expectedRoot2, err := computeEthReferenceTrieHash(vals2)
	if err != nil {
		t.Fatal(err)
	}
	err = addTrie(txTries, expectedRoot2, vals2)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != expectedRoot2 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot2, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot2 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot2, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	vals3 := GetTransactions3()
	expectedRoot3, err := computeEthReferenceTrieHash(vals3)
	if err != nil {
		t.Fatal(err)
	}
	err = addTrie(txTries, expectedRoot3, vals3)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != expectedRoot3 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot3, txTries.txRoots[2])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot3 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot3, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	keyToRetrieve, err := rlp.EncodeToBytes(uint(0))
	if err != nil {
		t.Fatal(err)
	}

	_, err = txTries.RetrieveProof(expectedRoot1, keyToRetrieve)
	if err == nil {
		t.Fatalf("able to retrieve proof from deleted trie")
	}

	_, err = txTries.RetrieveProof(expectedRoot2, keyToRetrieve)
	if err == nil {
		t.Fatalf("able to retrieve proof from deleted trie")
	}

	proofDb3, err := txTries.RetrieveProof(expectedRoot3, keyToRetrieve)
	if err != nil {
		t.Fatal(err)
	}

	exists3, err := VerifyProof(expectedRoot3, keyToRetrieve, proofDb3)
	if err != nil {
		t.Fatal(err)
	}

	if exists3 != true {
		t.Fatalf("not able to verify retrieved proof!")
	}
}

func TestRetrieveEncodedProof(t *testing.T) {
	txTries := NewTxTries()

	vals1 := GetTransactions1()
	expectedRoot1, err := computeEthReferenceTrieHash(vals1)
	if err != nil {
		t.Fatal(err)
	}
	err = addTrie(txTries, expectedRoot1, vals1)
	if err != nil {
		t.Fatal(err)
	}

	key := uint(0)

	keyRlp, err := rlp.EncodeToBytes(key)

	if err != nil {
		t.Error("failed to encode key")
	}

	_, err = txTries.RetrieveEncodedProof(expectedRoot1, keyRlp)

	if err != nil {
		t.Error("unable to rerieve proof")
	}
}
