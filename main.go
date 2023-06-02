package main

import (
	"encoding/hex"
	"felica_cards/felica_cards"
	"fmt"
	"github.com/ebfe/scard"
	"strings"
)

func main() {

	rc := []byte{0x9a, 0x82, 0xef, 0x5a, 0xdd, 0x47, 0xe2, 0x51, 0xc9, 0x48, 0x74, 0x8e, 0x25, 0x29, 0x55, 0x96}
	information := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	sessionKey := felica_cards.SessionKeyGenerationMAC(
		rc,
		[]byte{0x6d, 0x2d, 0xd8, 0x40, 0x9b, 0xbe, 0x2d, 0x5b, 0x4a, 0x91, 0x56, 0xd8, 0x8f, 0x54, 0x12, 0x6f})

	mac := felica_cards.MACKeyGeneration(
		information,
		rc,
		sessionKey)

	println(mac)

	macA := felica_cards.MACAReadGeneration(
		information,
		rc,
		sessionKey,
		[]byte{0x05, 0x00, 0x91, 0x00, 0xff, 0xff, 0xff, 0xff},
	)
	println(macA)

	pcscContext, err := scard.EstablishContext()
	handleError(err)

	defer pcscContext.Release()

	readerList, err := pcscContext.ListReaders()
	handleError(err)

	pcscInfo := felica_cards.NewScreenInfoRawPCSC()
	pcscInfo.ReaderList = readerList

	displayShow := felica_cards.GetWindow(pcscInfo)
	selectedReader := ""

	newChan := make(chan *scard.Card)

	if len(readerList) == 1 {
		selectedReader = readerList[0]
		pcscInfo.AutoChooseReader(readerList[0])
		go waitForCard(pcscContext, selectedReader, newChan)
	}

	go func() {
		for {
			resType, resContent := displayShow.GetEvent()

			switch resType {
			case felica_cards.NewReaderSelect:
				if selectedReader != resContent {
					selectedReader = resContent
					go waitForCard(pcscContext, selectedReader, newChan)
				}
			case felica_cards.SubmitPCSCCode:
				card := <-newChan
				inputText, err := pcscInfo.PcscInput.Get()
				newContent := strings.ReplaceAll(inputText, " ", "")
				bytes, err := hex.DecodeString(newContent)
				if err != nil {
					fmt.Errorf("failed to decode hex: %w", err)
				}
				rsp, err := card.Transmit(bytes)
				handleError(err)

				var outputBuilder strings.Builder

				for i := 0; i < len(rsp); i++ {
					fmt.Fprintf(&outputBuilder, "%02x ", rsp[i])
					if (i+1)%16 == 0 {
						fmt.Fprint(&outputBuilder, "\n")
					}
				}

				pcscInfo.UpdateOutput(outputBuilder.String())

				card.Disconnect(scard.LeaveCard)

			}
		}
	}()

	displayShow.StartWindow()

}

func waitForCard(context *scard.Context, selectedReader string, newCard chan *scard.Card) {
	for {
		card, err := context.Connect(selectedReader, scard.ShareShared, scard.ProtocolAny)
		handleError(err)

		if card != nil {

			newCard <- card

			//var cmdRead = []byte{0xff, 0xb0, 0x00, 0x00, 0x10}
		}
	}
}

func readCardInfo(context *scard.Context, card *scard.Card) {

}

func handleError(e error) {
	if e != nil {
		fmt.Println("Error:", e)
	}
}

//func main() {
//
//}
