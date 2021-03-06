// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package txtrie

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
)

func newTestFullNode(v []byte) []interface{} {
	var fullNodeData []interface{}
	for i := 0; i < 16; i++ {
		k := bytes.Repeat([]byte{byte(i + 1)}, 32)
		fullNodeData = append(fullNodeData, k)
	}
	fullNodeData = append(fullNodeData, v)
	return fullNodeData
}

func TestDecodeNestedNode(t *testing.T) {
	fullNodeData := newTestFullNode([]byte("fullnode"))

	var data [][]byte
	for i := 0; i < 16; i++ {
		data = append(data, nil)
	}
	data = append(data, []byte("subnode"))
	fullNodeData[15] = data

	buf := bytes.NewBuffer([]byte{})
	if err := rlp.Encode(buf, fullNodeData); err != nil {
		t.Fatalf("unable to rlp encode node %v", err)
	}

	if _, err := decodeNode([]byte("testdecode"), buf.Bytes()); err != nil {
		t.Fatalf("decode nested full node err: %v", err)
	}
}

func TestDecodeFullNodeWrongSizeChild(t *testing.T) {
	fullNodeData := newTestFullNode([]byte("wrongsizechild"))
	fullNodeData[0] = []byte("00")
	buf := bytes.NewBuffer([]byte{})
	if err := rlp.Encode(buf, fullNodeData); err != nil {
		t.Fatalf("unable to rlp encode node %v", err)
	}

	_, err := decodeNode([]byte("testdecode"), buf.Bytes())
	if _, ok := err.(*decodeError); !ok {
		t.Fatalf("decodeNode returned wrong err: %v", err)
	}
}

func TestDecodeFullNodeWrongNestedFullNode(t *testing.T) {
	fullNodeData := newTestFullNode([]byte("fullnode"))

	var data [][]byte
	for i := 0; i < 16; i++ {
		data = append(data, []byte("123456"))
	}
	data = append(data, []byte("subnode"))
	fullNodeData[15] = data

	buf := bytes.NewBuffer([]byte{})
	if err := rlp.Encode(buf, fullNodeData); err != nil {
		t.Fatalf("unable to rlp encode node %v", err)
	}

	_, err := decodeNode([]byte("testdecode"), buf.Bytes())
	if _, ok := err.(*decodeError); !ok {
		t.Fatalf("decodeNode returned wrong err: %v", err)
	}
}

func TestDecodeFullNode(t *testing.T) {
	fullNodeData := newTestFullNode([]byte("decodefullnode"))
	buf := bytes.NewBuffer([]byte{})
	if err := rlp.Encode(buf, fullNodeData); err != nil {
		t.Fatalf("unable to rlp encode node %v", err)
	}

	_, err := decodeNode([]byte("testdecode"), buf.Bytes())
	if err != nil {
		t.Fatalf("decode full node err: %v", err)
	}
}
