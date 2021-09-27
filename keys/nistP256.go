package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"

	"github.com/completium/go-tezos/v4/internal/crypto"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
)

var _ iCurve = &nistP256Curve{}

type nistP256Curve struct{}

func (n *nistP256Curve) addressPrefix() []byte {
	return []byte{6, 161, 164}
}

func (n *nistP256Curve) publicKeyPrefix() []byte {
	return []byte{3, 178, 139, 127}
}

func (n *nistP256Curve) privateKeyPrefix() []byte {
	return []byte{16, 81, 238, 189}
}

func (n *nistP256Curve) signaturePrefix() []byte {
	return []byte{54, 240, 44, 52}
}

func (n *nistP256Curve) getECKind() ECKind {
	return NistP256
}

func (n *nistP256Curve) getPrivateKey(v []byte) []byte {
	return v[:32]
}

func (n *nistP256Curve) getPublicKey(privateKey []byte) ([]byte, error) {
	var privKey ecdsa.PrivateKey
	privKey.D = new(big.Int).SetBytes(privateKey)
	privKey.PublicKey.Curve = elliptic.P256()
	privKey.PublicKey.X, privKey.PublicKey.Y = privKey.PublicKey.Curve.ScalarBaseMult(privKey.D.Bytes())

	var pref []byte
	if privKey.PublicKey.Y.Bytes()[31]%2 == 0 {
		pref = []byte{2}
	} else {
		pref = []byte{3}
	}

	// 32 padded 0's
	pad := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	pad = append(pad, privKey.PublicKey.X.Bytes()...)

	return append(pref, pad[len(pad)-32:]...), nil
}

func (n *nistP256Curve) sign(msg []byte, privateKey []byte) (Signature, error) {
	hash, err := blake2b.New(32, []byte{})
	if err != nil {
		return Signature{}, err
	}

	i, err := hash.Write(msg)
	if err != nil {
		return Signature{}, errors.Wrap(err, "failed to sign operation bytes")
	}
	if i != len(msg) {
		return Signature{}, errors.Errorf("failed to sign operation: generic hash length %d does not match bytes length %d", i, len(msg))
	}

	var privKey ecdsa.PrivateKey
	privKey.D = new(big.Int).SetBytes(privateKey)
	privKey.PublicKey.Curve = elliptic.P256()
	privKey.PublicKey.X, privKey.PublicKey.Y = privKey.PublicKey.Curve.ScalarBaseMult(privKey.D.Bytes())

	r, ss, err := ecdsa.Sign(rand.Reader, &privKey, hash.Sum([]byte{}))
	if err != nil {
		return Signature{}, err
	}

	signature := append(r.Bytes(), ss.Bytes()...)

	return Signature{
		Bytes:  signature,
		prefix: n.signaturePrefix(),
	}, nil
}

func (n *nistP256Curve) checkSignature(pubKey []byte, hash []byte, signature []byte) (bool, error) {
	rb := signature[0:32]
	sb := signature[32:64]

	r := new(big.Int)
	r.SetBytes(rb)

	s := new(big.Int)
	s.SetBytes(sb)

	sk := crypto.B58cdecode("p2sk2mJNRYqs3UXJzzF44Ym6jk38RVDPVSuLCfNd5ShE5zyVdu8Au9", n.privateKeyPrefix())
	privKey, err := ethcrypto.ToECDSA(sk)
	if err != nil {
		return false, err
	}
	pk := privKey.PublicKey

	// fmt.Printf("ref: X: %i\n", pk.X)
	// fmt.Printf("ref: Y: %i\n", pk.Y)
	// fmt.Printf("in: %s", hex.EncodeToString(pubKey))
	// pkb := crypto.B58cdecode("sppk7b4TURq2T9rhPLFaSz6mkBCzKzfiBjctQSMorvLD5GSgCduvKuf", sec.publicKeyPrefix())

	// pkm, err := ethcrypto.UnmarshalPubkey(pubKey)
	// if err != nil {
	// 	return false, err
	// }
	// fmt.Printf("my : X: %i\n", pkm.X)
	// fmt.Printf("my : Y: %i\n", pkm.Y)

	// if err != nil {
	// 	return false, err
	// }

	return ecdsa.Verify(&pk, hash, r, s), nil
}
