package felica_cards

import (
	"crypto/cipher"
	"crypto/des"
)

type FelicaCardManagement struct {
	selectedReader string
}

func BoogeyFunc(entered bool) {}

func SessionKeyGenerationMAC(rc []byte, ck []byte) []byte {

	rcRev1 := reverseBytes(rc[0:8])
	rcRev2 := reverseBytes(rc[8:])

	ckRev1 := reverseBytes(ck[0:8])
	ckRev2 := reverseBytes(ck[8:])

	ck1Block, _ := des.NewCipher(ckRev1)
	ck2Block, _ := des.NewCipher(ckRev2)

	twoKeyTripleDES(ck1Block, ck2Block, rcRev1)

	for i := range rcRev2 {
		rcRev2[i] = rcRev2[i] ^ rcRev1[i]
	}

	twoKeyTripleDES(ck1Block, ck2Block, rcRev2)

	newrc1 := reverseBytes(rcRev1)
	newrc2 := reverseBytes(rcRev2)

	res := append(newrc1, newrc2...)

	return res
}

func MACKeyGeneration(info []byte, rc []byte, sk []byte) []byte {

	infoRev1 := reverseBytes(info[0:8])
	infoRev2 := reverseBytes(info[8:])

	rcRev1 := reverseBytes(rc[0:8])

	skRev1 := reverseBytes(sk[0:8])
	skRev2 := reverseBytes(sk[8:])

	sk1Block, _ := des.NewCipher(skRev1)
	sk2Block, _ := des.NewCipher(skRev2)

	for i := range rcRev1 {
		infoRev1[i] = infoRev1[i] ^ rcRev1[i]
	}

	twoKeyTripleDES(sk1Block, sk2Block, infoRev1)

	for i := range infoRev2 {
		infoRev2[i] = infoRev2[i] ^ infoRev1[i]
	}

	twoKeyTripleDES(sk1Block, sk2Block, infoRev2)

	return reverseBytes(infoRev2)
}

func MACAReadGeneration(info []byte, rc []byte, sk []byte, blockList []byte) []byte {

	rcRev1 := reverseBytes(rc[0:8])
	blockListRev := reverseBytes(blockList)

	skRev1 := reverseBytes(sk[0:8])
	skRev2 := reverseBytes(sk[8:])

	sk1Block, _ := des.NewCipher(skRev1)
	sk2Block, _ := des.NewCipher(skRev2)

	for i := range blockListRev {
		blockListRev[i] = blockListRev[i] ^ rcRev1[i]
	}

	twoKeyTripleDES(sk1Block, sk2Block, blockListRev)

	var infoRev []byte
	var oldRev []byte

	for i := 8; i <= len(info); i += 8 {

		if i == 8 {
			infoRev = reverseBytes(info[i-8 : i])
			for j := range blockListRev {
				infoRev[j] = blockListRev[j] ^ infoRev[j]
			}
		} else {
			infoRev = reverseBytes(info[i-8 : i])
			for j := range infoRev {
				infoRev[j] = oldRev[j] ^ infoRev[j]
			}
		}

		twoKeyTripleDES(sk1Block, sk2Block, infoRev)
		oldRev = infoRev

	}

	return reverseBytes(infoRev)
}

func MACAWriteGeneration() {

}

func twoKeyTripleDES(block1 cipher.Block, block2 cipher.Block, in []byte) {
	block1.Encrypt(in, in)
	block2.Decrypt(in, in)
	block1.Encrypt(in, in)
}

func reverseBytes(list []byte) []byte {
	output := make([]byte, len(list))
	copy(output[:], list)

	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		output[i], output[j] = output[j], output[i]
	}
	return output
}
