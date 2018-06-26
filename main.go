package main

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"log"
	"math"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func main() {

	// TODO: read in file

	// TODO: fingerprint and confirm file

	// TODO: foreach key

	gt := k.Config.Now()

	config := k.Config

	uid := (*k.UserIds)[0].ToPacket()
	if uid == nil {
		return
	}

	primaryKey, err := rsa.GenerateKey(config.Random(), k.Length)
	if err != nil {
		return
	}

	e := &openpgp.Entity{
		PrimaryKey: packet.NewRSAPublicKey(gt, &primaryKey.PublicKey),
		PrivateKey: packet.NewRSAPrivateKey(gt, primaryKey),
		Identities: make(map[string]*openpgp.Identity),
	}

	e.Identities[uid.Id] = &openpgp.Identity{
		Name:          uid.Name,
		UserId:        uid,
		SelfSignature: k.GenerateSelfSig(gt, &e.PrimaryKey.KeyId),
	}

	e.Subkeys, err = k.GenerateSubKeys(gt, e)
	if err != nil {
		return
	}

	signingKey := packet.NewRSAPrivateKey(gt, primaryKey)

	for _, id := range e.Identities {
		err := id.SelfSignature.SignUserId(uid.Id, e.PrimaryKey, signingKey, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Sign subkeys
	for _, subkey := range e.Subkeys {
		err := subkey.Sig.SignKey(subkey.PublicKey, signingKey, nil)
		if err != nil {
			log.Fatal(err)
		}
	}

	buffer := &bytes.Buffer{}

	w, err := armor.Encode(buffer, openpgp.PublicKeyType, map[string]string{})

	e.Serialize(w)
	w.Close()

	r, err := armor.Decode(strings.NewReader(buffer.String()))
	fromReader := packet.NewReader(r.Body)
	_, err = openpgp.ReadEntity(fromReader)
	if err != nil {
		log.Fatal(err)
	}
}

func groupLine(s []string, groupLen int) string {
	l := len(s)
	ss := "\t"
	for i := 0; i < l; i++ {
		ss += s[i] + " "
		m := math.Mod(float64(i), float64(groupLen))
		if m == float64(groupLen-1) {
			ss += "\n\t"
		}
	}
	return ss
}
