package main

import (
	"context"
	"felica_cards/felica_cards"
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
				} else {
					memberUICancel()
					memberUIContext, memberUICancel = context.WithCancel(context.Background())
					felicaManagement.ReaderCommand <- felica_cards.CancelWait
					felicaManagement.CancelWait()
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
				inputText, _ := pcscInfoScreen.PcscInput.Get()
				result := felicaManagement.WriteRawPCSC(inputText)
				pcscInfoScreen.UpdateOutput(result)
			}
		}
	}()

	displayShow.StartWindow()
	memberUICancel()
	felicaManagement.Release()

}

//func main() {
//
//}
