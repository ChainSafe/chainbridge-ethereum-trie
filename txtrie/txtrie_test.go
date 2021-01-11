// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package txtrie

import (
	"fmt"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

var (
	defaultTriesToStore = 3
	emptyHash           = common.HexToHash("")
)

func createNewTxTries(numHistoricalTries int) *TxTries {
	return NewTxTries(numHistoricalTries)
}

func createTempDB() *leveldb.Database {

	diskdb, err := leveldb.New("./temp-database", 256, 0, "")
	if err != nil {
		panic(fmt.Sprintf("unable to create testing database: %v", err))
	}
	return diskdb
}

func deleteTempDB() error {
	err := os.RemoveAll("./temp-database")

	if err != nil {
		return err
	}

	return nil
}

func createReferenceDB() *leveldb.Database {

	diskdb, err := leveldb.New("./reference-database", 256, 0, "")
	if err != nil {
		panic(fmt.Sprintf("unable to create reference database: %v", err))
	}
	return diskdb
}

func deleteReferenceDB() error {
	err := os.RemoveAll("./reference-database")

	if err != nil {
		return err
	}

	return nil
}

func addTrie(txTries *TxTries, root common.Hash, transactions types.Transactions, db *leveldb.Database) error {

	if db == nil {
		db = createTempDB()
	}

	if transactions == nil {
		transactions = types.Transactions{}
	}

	err := txTries.AddNewTrie(root, transactions, db)

	if err != nil {
		return err
	}

	return nil

}

func computeEthReferenceTrieHash(transactions types.Transactions) (common.Hash, error) {
	db := createReferenceDB()
	newTrie, err := trie.New(emptyRoot, trie.NewDatabaseWithCache(db, 0, ""))
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

func TestEmptyTxTries(t *testing.T) {
	txTries := createNewTxTries(defaultTriesToStore)

	if txTries.triesToStore != defaultTriesToStore {
		t.Fatalf("tries to store not set properly, expected: %x, got: %x", defaultTriesToStore, txTries.triesToStore)
	}
}

func TestAddEmptyTrie(t *testing.T) {
	txTries := createNewTxTries(defaultTriesToStore)
	err := addTrie(txTries, emptyRoot, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != emptyRoot {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", emptyRoot, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != emptyRoot {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", emptyRoot, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	if deleteTempDB() != nil {
		t.Fatalf("unable to clear testing database")
	}

}

func TestAddEmptyTrieRetrieveProof_Fails(t *testing.T) {
	txTries := createNewTxTries(defaultTriesToStore)
	err := addTrie(txTries, emptyRoot, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != emptyRoot {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", emptyRoot, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != emptyRoot {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", emptyRoot, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	if deleteTempDB() != nil {
		t.Fatalf("unable to clear testing database")
	}
}

func TestAddSingleTrieUpdate(t *testing.T) {
	vals := GetTransactions1()
	expectedRoot, err := computeEthReferenceTrieHash(vals)
	if err != nil {
		t.Fatal(err)
	}

	txTries := createNewTxTries(defaultTriesToStore)
	err = addTrie(txTries, expectedRoot, vals, nil)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != expectedRoot {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	if deleteTempDB() != nil {
		t.Fatalf("unable to clear testing database")
	}

	if deleteReferenceDB() != nil {
		t.Fatalf("unable to clear reference database")
	}
}

func TestAddSingleTrieRetrieveProof(t *testing.T) {
	vals := GetTransactions1()
	expectedRoot, err := computeEthReferenceTrieHash(vals)
	if err != nil {
		t.Fatal(err)
	}

	txTries := createNewTxTries(defaultTriesToStore)
	err = addTrie(txTries, expectedRoot, vals, nil)
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

	if deleteTempDB() != nil {
		t.Fatalf("unable to clear testing database")
	}

	if deleteReferenceDB() != nil {
		t.Fatalf("unable to clear reference database")
	}
}

func TestAddMultipleTries(t *testing.T) {
	txTries := createNewTxTries(defaultTriesToStore)
	db := createTempDB()

	vals1 := GetTransactions1()
	expectedRoot1, err := computeEthReferenceTrieHash(vals1)
	if err != nil {
		t.Fatal(err)
	}

	err = addTrie(txTries, expectedRoot1, vals1, db)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != expectedRoot1 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot1, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot1 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot1, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	if deleteReferenceDB() != nil {
		t.Fatalf("unable to clear reference database")
	}

	vals2 := GetTransactions2()
	expectedRoot2, err := computeEthReferenceTrieHash(vals2)
	if err != nil {
		t.Fatal(err)
	}

	err = addTrie(txTries, expectedRoot2, vals2, db)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[1] != expectedRoot2 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot2, txTries.txRoots[1])
	}

	if txTries.txTries[txTries.txRoots[1]].Hash() != expectedRoot2 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot2, txTries.txTries[txTries.txRoots[1]].Hash())
	}

	if deleteReferenceDB() != nil {
		t.Fatalf("unable to clear reference database")
	}

	vals3 := GetTransactions3()
	expectedRoot3, err := computeEthReferenceTrieHash(vals3)
	if err != nil {
		t.Fatal(err)
	}

	err = addTrie(txTries, expectedRoot3, vals3, db)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[2] != expectedRoot3 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot3, txTries.txRoots[2])
	}

	if txTries.txTries[txTries.txRoots[2]].Hash() != expectedRoot3 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot3, txTries.txTries[txTries.txRoots[2]].Hash())
	}

	err = addTrie(txTries, expectedRoot1, vals1, db)
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

	if deleteTempDB() != nil {
		t.Fatalf("unable to clear testing database")
	}

	if deleteReferenceDB() != nil {
		t.Fatalf("unable to clear reference database")
	}
}

func TestAddMultipleTriesRetrieveProof(t *testing.T) {
	txTries := createNewTxTries(defaultTriesToStore)
	db := createTempDB()

	vals1 := GetTransactions1()
	expectedRoot1, err := computeEthReferenceTrieHash(vals1)
	if err != nil {
		t.Fatal(err)
	}

	err = addTrie(txTries, expectedRoot1, vals1, db)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != expectedRoot1 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot1, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot1 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot1, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	if deleteReferenceDB() != nil {
		t.Fatalf("unable to clear reference database")
	}

	vals2 := GetTransactions2()
	expectedRoot2, err := computeEthReferenceTrieHash(vals2)
	if err != nil {
		t.Fatal(err)
	}

	err = addTrie(txTries, expectedRoot2, vals2, db)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[1] != expectedRoot2 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot2, txTries.txRoots[1])
	}

	if txTries.txTries[txTries.txRoots[1]].Hash() != expectedRoot2 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot2, txTries.txTries[txTries.txRoots[1]].Hash())
	}

	if deleteReferenceDB() != nil {
		t.Fatalf("unable to clear reference database")
	}

	vals3 := GetTransactions3()
	expectedRoot3, err := computeEthReferenceTrieHash(vals3)
	if err != nil {
		t.Fatal(err)
	}

	err = addTrie(txTries, expectedRoot3, vals3, db)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[2] != expectedRoot3 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot3, txTries.txRoots[2])
	}

	if txTries.txTries[txTries.txRoots[2]].Hash() != expectedRoot3 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot3, txTries.txTries[txTries.txRoots[2]].Hash())
	}

	err = addTrie(txTries, expectedRoot1, vals1, db)
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

	if deleteTempDB() != nil {
		t.Fatalf("unable to clear testing database")
	}

	if deleteReferenceDB() != nil {
		t.Fatalf("unable to clear reference database")
	}
}

func TestRetrieveProofDeletedTrie_Fails(t *testing.T) {
	txTries := createNewTxTries(1)
	db := createTempDB()

	vals1 := GetTransactions1()
	expectedRoot1, err := computeEthReferenceTrieHash(vals1)
	if err != nil {
		t.Fatal(err)
	}
	err = addTrie(txTries, expectedRoot1, vals1, db)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != expectedRoot1 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot1, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot1 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot1, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	if deleteReferenceDB() != nil {
		t.Fatalf("unable to clear reference database")
	}

	vals2 := GetTransactions2()
	expectedRoot2, err := computeEthReferenceTrieHash(vals2)
	if err != nil {
		t.Fatal(err)
	}
	err = addTrie(txTries, expectedRoot2, vals2, db)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != expectedRoot2 {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot2, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot2 {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot2, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	if deleteReferenceDB() != nil {
		t.Fatalf("unable to clear reference database")
	}

	vals3 := GetTransactions3()
	expectedRoot3, err := computeEthReferenceTrieHash(vals3)
	if err != nil {
		t.Fatal(err)
	}
	err = addTrie(txTries, expectedRoot3, vals3, db)
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

	if deleteTempDB() != nil {
		t.Fatalf("unable to clear testing database")
	}

	if deleteReferenceDB() != nil {
		t.Fatalf("unable to clear reference database")
	}
}

func TestRetrieveEncodedProof(t *testing.T) {
	txTries := createNewTxTries(1)
	db := createTempDB()

	vals1 := GetTransactions1()
	expectedRoot1, err := computeEthReferenceTrieHash(vals1)
	if err != nil {
		t.Fatal(err)
	}
	err = addTrie(txTries, expectedRoot1, vals1, db)
	if err != nil {
		t.Fatal(err)
	}

	keyRlp, err := rlp.EncodeToBytes(0)

	if err != nil {
		t.Error("failed to encode key")
	}

	_, err = txTries.RetrieveEncodedProof(expectedRoot1, keyRlp)

	if err != nil {
		t.Error("unable to rerieve proof")
	}
}
