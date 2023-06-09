package felica_cards

import (
	"context"
	"encoding/hex"
	"github.com/ebfe/scard"
	"os"
)

type CardCommand int

const (
	ReadMember   = CardCommand(1)
	RawPCSC      = CardCommand(2)
	ChangeReader = CardCommand(100)
)

type CardReaderManager struct {
	context        *scard.Context
	readerList     []string
	selectedReader string

	readerCommand chan CardCommand
	waitContext   context.Context
	waitCancel    context.CancelFunc

	MemberChannel chan MemberInfo
}

type MemberInfo struct {
	CardID   string
	memberId string
	name     string
}

func SetupInfo() CardReaderManager {
	establishContext, err := scard.EstablishContext()
	if err != nil {
		println("Error at Establishing Context")
		os.Exit(1)
	}

	readers, err := establishContext.ListReaders()
	if err != nil {
		println("Error at getting Readers")
	}

	selectedReader := ""
	if len(readers) == 1 {
		selectedReader = readers[0]
	}

	waitContext, waitCancel := context.WithCancel(context.Background())

	return CardReaderManager{
		context:        establishContext,
		readerList:     readers,
		selectedReader: selectedReader,

		readerCommand: make(chan CardCommand),
		waitContext:   waitContext,
		waitCancel:    waitCancel,

		MemberChannel: make(chan MemberInfo),
	}

}

func (cardMan *CardReaderManager) GetContext() *scard.Context {
	return cardMan.context
}

func (cardMan *CardReaderManager) Release() {
	cardMan.waitCancel()
	err := cardMan.context.Cancel()
	if err != nil {
		return
	}

	err = cardMan.context.Release()
	if err != nil {
		return
	}
}

func (cardMan *CardReaderManager) GetReader() string {
	return cardMan.selectedReader
}

func (cardMan *CardReaderManager) GetReaderList() []string {
	return cardMan.readerList
}

func (cardMan *CardReaderManager) ChangeReader(newReader string) bool {
	if newReader != cardMan.selectedReader {
		cardMan.selectedReader = newReader
		err := cardMan.context.Cancel()
		if err != nil {
			return false
		}
		cardMan.waitCancel()
		return true
	}
	return false
}

func (cardMan *CardReaderManager) WaitForCard(cardCommand chan CardCommand) {

	currentReaderState := []scard.ReaderState{
		{
			Reader:       cardMan.selectedReader,
			CurrentState: scard.StateUnaware,
		},
	}

	currentStatus := ReadMember

	for {
		select {
		case <-cardMan.waitContext.Done():
			return
		case x := <-cardCommand:
			currentStatus = x
		default:
			switch currentStatus {
			case ReadMember:
				err := cardMan.context.GetStatusChange(currentReaderState, 1<<63-1)
				if err != nil {
					return
				}
				eventState := currentReaderState[0].EventState & 0x00F0
				if eventState == scard.StatePresent {
					card, err := cardMan.context.Connect(cardMan.selectedReader, scard.ShareShared, scard.ProtocolAny)
					if err != nil {
						//Fill this in if needed
					}
					memInfo := MemberInfo{
						CardID: readCardID(card),
					}

					cardMan.MemberChannel <- memInfo

				}
			}

			currentReaderState[0].CurrentState = currentReaderState[0].EventState & 0x00F0
		}
	}

	//for {
	//
	//	card, err := context.Connect(selectedReader, scard.ShareShared, scard.ProtocolAny)
	//	if err != nil {
	//		//Fill this in if needed
	//	}
	//
	//	if card != nil {
	//
	//		newCard <- card
	//
	//
	//	} else {
	//
	//	}
	//}
}

func readCardID(card *scard.Card) string {
	cardInfo := card
	getID := []byte{0xff, 0xb0, 0x80, 0x01, 0x02, 0x80, 0x82, 0x80, 0x00, 0x20}
	output, _ := cardInfo.Transmit(getID)
	if output != nil {
		return hex.EncodeToString(output[:len(getID)-2])
	}
	return ""
}
