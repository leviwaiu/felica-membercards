package main

import (
	"context"
	"felica_cards/felica_cards"
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
	readToggle := false

	go func() {
		for {
			resType, resContent := displayShow.GetEvent()

			switch resType {
			case felica_cards.UISelectReader:
				if felicaManagement.ChangeReader(resContent) {
					go felicaManagement.WaitForCard(cardCommand)
				}
			case felica_cards.UIWritePCSC:
				inputText, _ := pcscInfoScreen.PcscInput.Get()
				result := felicaManagement.WriteRawPCSC(inputText)
				pcscInfoScreen.UpdateOutput(result)
			case felica_cards.UIWriteCard:
			case felica_cards.UIReadCard:
				result := felicaManagement.ReadCardInfo()
				memberInfoScreen.UpdateMemberInfo(result)
			case felica_cards.UIReadToggle:
				if !readToggle {
					felicaManagement.ReaderCommand <- felica_cards.ReadMember

					go func() {
						for {
							select {
							case <-memberUIContext.Done():
								return
							case memberInfo := <-felicaManagement.MemberChannel:
								memberInfoScreen.UpdateMemberInfo(memberInfo)
							}
						}
					}()
					readToggle = true
				} else {
					memberUICancel()
					memberUIContext, memberUICancel = context.WithCancel(context.Background())
					felicaManagement.ReaderCommand <- felica_cards.CancelWait
					felicaManagement.CancelWait()
					readToggle = false
				}
			}

		}
	}()

	displayShow.StartWindow()
	memberUICancel()
	felicaManagement.Release()

}
