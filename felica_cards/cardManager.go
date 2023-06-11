package felica_cards

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ebfe/scard"
	"os"
	"strings"
)

type CardCommand int

const (
	Idle         = CardCommand(0)
	ReadMember   = CardCommand(1)
	ChangeReader = CardCommand(100)
	CancelWait   = CardCommand(-1)
)

type ReadResult int

type CardReaderManager struct {
	context        *scard.Context
	readerList     []string
	selectedReader string

	ReaderCommand chan CardCommand
	waitContext   context.Context
	waitCancel    context.CancelFunc

	MemberChannel chan MemberInfo
	waitCanc      chan bool
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

		ReaderCommand: make(chan CardCommand),
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
		cardMan.waitContext, cardMan.waitCancel = context.WithCancel(context.Background())
		return true
	}
	return false
}

func (cardMan *CardReaderManager) CancelWait() {
	cardMan.context.Cancel()

}

func (cardMan *CardReaderManager) WaitForCard(cardCommand chan CardCommand) {
	//I feel like I have made this thing more complicated than it ever has to be

	currentReaderState := []scard.ReaderState{
		{
			Reader:       cardMan.selectedReader,
			CurrentState: scard.StateUnaware,
		},
	}

	currentStatus := ReadMember
	running := false
	runningChan := make(chan bool)

	for {
		select {
		case x := <-cardMan.ReaderCommand:
			currentStatus = x
		case newCommand := <-runningChan:

			running = newCommand
		default:
			switch currentStatus {
			case ChangeReader:

			case CancelWait:
				running = false
				currentStatus = Idle
			case ReadMember:
				for !running {
					go func() {
						select {
						case <-cardMan.waitContext.Done():
							running = false
							return
						default:
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

								id, success := readCardID(card)
								if !success {

								}

								memInfo := MemberInfo{
									CardID: id,
								}

								cardMan.MemberChannel <- memInfo
								card.Disconnect(scard.LeaveCard)
							}
							runningChan <- false
						}
					}()
					running = true
				}
			}
			currentReaderState[0].CurrentState = currentReaderState[0].EventState & 0x00F0
		}
	}
}

func readCardID(card *scard.Card) (string, bool) {
	getID := []byte{0xff, 0xb0, 0x80, 0x01, 0x02, 0x80, 0x82, 0x80, 0x00, 0x20}
	output, _ := card.Transmit(getID)

	if bytes.Equal(output, []byte{0x69, 0x85}) {
		return "ERROR: Card Cannot Be Read", false
	}

	if output != nil {
		return hex.EncodeToString(output[:len(getID)-2]), true
	}
	return "", false
}

func readCardInfo(card *scard.Card) {
	getMemberInfo := []byte{0xff, 0xb0, 0x80, 0x03, 0x06, 0x80, 0x00, 0x80, 0x01, 0x80, 0x02, 0x30}
	output, _ := card.Transmit(getMemberInfo)

	if output != nil {

	}

}

func (cardMan *CardReaderManager) WriteRawPCSC(inputText string) string {

	card, err := cardMan.context.Connect(cardMan.selectedReader, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		return ""
	}

	newContent := strings.ReplaceAll(inputText, " ", "")
	bytes, err := hex.DecodeString(newContent)
	if err != nil {
		fmt.Errorf("failed to decode hex: %w", err)
	}
	rsp, _ := card.Transmit(bytes)

	var outputBuilder strings.Builder

	for i := 0; i < len(rsp); i++ {
		fmt.Fprintf(&outputBuilder, "%02x ", rsp[i])
		if (i+1)%16 == 0 {
			fmt.Fprint(&outputBuilder, "\n")
		}
	}

	card.Disconnect(scard.LeaveCard)
	return outputBuilder.String()
}
