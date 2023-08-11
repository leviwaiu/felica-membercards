package felica_cards

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
)

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

func MACAWriteGeneration(info []byte, rc []byte, sk []byte, wcntList []byte) []byte {

	//Leaving this separate to add checks if needed
	if len(info) > 16 {

	}

	return MACAReadGeneration(info, rc, sk, wcntList)

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

func writeRandomScratch() {
	//TODO
}

func generateCardKey(master []byte, cardId []byte) []byte {
	newBlock, _ := aes.NewCipher(master)

	destination := make([]byte, len(cardId))
	newBlock.Encrypt(destination, cardId)

	return destination
}

//rc := []byte{0x9a, 0x82, 0xef, 0x5a, 0xdd, 0x47, 0xe2, 0x51, 0xc9, 0x48, 0x74, 0x8e, 0x25, 0x29, 0x55, 0x96}

//information := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
//
//sessionKey := felica_cards.SessionKeyGenerationMAC(
//	rc,
//	[]byte{0x6d, 0x2d, 0xd8, 0x40, 0x9b, 0xbe, 0x2d, 0x5b, 0x4a, 0x91, 0x56, 0xd8, 0x8f, 0x54, 0x12, 0x6f})
//
//mac := felica_cards.MACKeyGeneration(
//	information,
//	rc,
//	sessionKey)
//
//println(mac)
//
//macA := felica_cards.MACAReadGeneration(
//	information,
//	rc,
//	sessionKey,
//	[]byte{0x05, 0x00, 0x91, 0x00, 0xff, 0xff, 0xff, 0xff},
//)
//println(macA)
