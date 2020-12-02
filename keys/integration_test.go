package keys

import (
	"encoding/hex"
	"os"
	"strconv"
	"testing"

	"github.com/goat-systems/go-tezos/v3/forge"
	"github.com/goat-systems/go-tezos/v3/internal/testutils"
	"github.com/goat-systems/go-tezos/v3/rpc"
	"github.com/stretchr/testify/assert"
)

func Test_OperationWithKey(t *testing.T) {
	cases := []struct {
		name    string
		input   NewKeyInput
		wantErr bool
	}{
		{
			"is successful Ed25519",
			NewKeyInput{
				EncodedString: "edsk2oJWw5CX7Fh3g8QDqtK9CmrvRDDDSeHAPvWPnm7CwD3RfQ1KbK",
				Kind:          Ed25519,
			},
			false,
		},
		{
			"is successful Secp256k1",
			NewKeyInput{
				EncodedString: "spsk1WCtWP1fEc4RaE63YK6oUEmbjLK2aTe7LevYSb9Z3zDdtq58wS",
				Kind:          Secp256k1,
			},
			false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rpchost := os.Getenv("GOTEZOS_TEST_RPC_HOST")
			r, _ := rpc.New(rpchost)

			key, err := NewKey(tt.input)
			testutils.CheckErr(t, tt.wantErr, "", err)

			head, _ := r.Head()

			counter, _ := r.Counter(rpc.CounterInput{
				Blockhash: head.Hash,
				Address:   key.PubKey.address,
			})

			transaction := rpc.Transaction{
				Kind:         rpc.TRANSACTION,
				Source:       key.PubKey.address,
				Fee:          "2941",
				Counter:      strconv.Itoa((counter + 1)),
				GasLimit:     "26283",
				Amount:       "1",
				StorageLimit: "0",
				Destination:  "tz1RomaiWJV3NFDZWTMVR2aEeHknsn3iF5Gi",
			}

			op, err := forge.Encode(head.Hash, transaction.ToContent())
			testutils.CheckErr(t, tt.wantErr, "", err)

			sig, err := key.Sign(SignInput{
				Message: op,
			})
			testutils.CheckErr(t, tt.wantErr, "", err)

			hexop, _ := hex.DecodeString(op)
			x := key.Verify(VerifyInput{
				BytesSignature: sig.Bytes,
				BytesData:      hexop,
			})
			assert.True(t, x)

			_, err = r.PreapplyOperations(rpc.PreapplyOperationsInput{
				Blockhash: head.Hash,
				Operations: []rpc.Operations{
					{
						Protocol:  head.Protocol,
						Branch:    head.Hash,
						Contents:  rpc.Contents{transaction.ToContent()},
						Signature: sig.ToBase58(),
					},
				},
			})
			testutils.CheckErr(t, tt.wantErr, "", err)
		})
	}
}
