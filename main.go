package main

import (
	"felica_cards/felica_cards"
	"fmt"
	"github.com/ebfe/scard"
	"golang.org/x/net/context"
	"time"
)

func main() {

	felicaManagement := felica_cards.SetupInfo()

	readerList := felicaManagement.GetReaderList()

	pcscInfoScreen := felica_cards.NewScreenInfoRawPCSC()
	pcscInfoScreen.ReaderList = readerList

	memberInfoScreen := felica_cards.NewScreenInfoMember()

	displayShow := felica_cards.GetWindow(pcscInfoScreen, memberInfoScreen)

	pcscInfoScreen.AutoChooseReader(readerList[0])

	cardCommand := make(chan felica_cards.CardCommand)
	go felicaManagement.WaitForCard(cardCommand)

	memberUIContext, memberUICancel := context.WithCancel(context.Background())
	go func() {
		previousCheck := false
		for {

			runInBackground := displayShow.CheckCardBackground()
			if runInBackground != previousCheck {
				if runInBackground {
					go func() {
						for {
							select {
							case <-memberUIContext.Done():
								return
							default:
								memberInfo := <-felicaManagement.MemberChannel
								memberInfoScreen.UpdateMemberInfo(memberInfo)
							}
						}
					}()
				} else {
					memberUICancel()
				}
				previousCheck = runInBackground
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	go func() {
		for {
			resType, resContent := displayShow.GetEvent()

			switch resType {
			case felica_cards.NewReaderSelect:
				if felicaManagement.ChangeReader(resContent) {
					go felicaManagement.WaitForCard(cardCommand)
				}
			case felica_cards.SubmitPCSCCode:
				//card := <-newChan
				//inputText, err := pcscInfoScreen.PcscInput.Get()
				//newContent := strings.ReplaceAll(inputText, " ", "")
				//bytes, err := hex.DecodeString(newContent)
				//if err != nil {
				//	fmt.Errorf("failed to decode hex: %w", err)
				//}
				//rsp, err := card.Transmit(bytes)
				//handleError(err)
				//
				//var outputBuilder strings.Builder
				//
				//for i := 0; i < len(rsp); i++ {
				//	fmt.Fprintf(&outputBuilder, "%02x ", rsp[i])
				//	if (i+1)%16 == 0 {
				//		fmt.Fprint(&outputBuilder, "\n")
				//	}
				//}
				//
				//pcscInfoScreen.UpdateOutput(outputBuilder.String())
				//
				//card.Disconnect(scard.LeaveCard)
			}
		}
	}()

	displayShow.StartWindow()
	memberUICancel()
	felicaManagement.Release()

}

func readCardInfo(card chan *scard.Card) {
	cardInfo := <-card

	getMemberInfo := []byte{0xff, 0xb0, 0x80, 0x03, 0x06, 0x80, 0x00, 0x80, 0x01, 0x80, 0x02, 0x30}
	output, _ := cardInfo.Transmit(getMemberInfo)
	if output != nil {

	}

}

func handleError(e error) {
	if e != nil {
		fmt.Println("Error:", e)
	}
}

//func main() {
//
//}
