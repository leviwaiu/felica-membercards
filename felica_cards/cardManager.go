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
								memInfo := readCardInfo(card)

								memInfo.CardID = id

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

func (cardMan *CardReaderManager) ReadCardInfo() MemberInfo {
	card, err := cardMan.context.Connect(cardMan.selectedReader, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		return MemberInfo{}
	}

	id, _ := readCardID(card)

	return MemberInfo{
		CardID: id,
	}
}

func (cardMan *CardReaderManager) WriteCardInfo(memInfo MemberInfo) {

}

func readCardID(card *scard.Card) (string, bool) {
	getID := []byte{0xff, 0xb0, 0x80, 0x01, 0x02, 0x80, 0x82, 0x00}
	output, _ := card.Transmit(getID)

	if bytes.Equal(output, []byte{0x69, 0x85}) {
		return "ERROR: Card Cannot Be Read", false
	}

	var idString strings.Builder

	idString.WriteString(hex.EncodeToString(output[:len(getID)-2]))
	for idString.Len() < 32 {
		idString.WriteRune('0')
	}

	if output != nil {
		return idString.String(), true
	}
	return "", false
}

func readCardInfo(card *scard.Card) MemberInfo {
	getMemberInfo := []byte{0xff, 0xb0, 0x80, 0x03, 0x06, 0x80, 0x00, 0x80, 0x01, 0x80, 0x02, 0x30}
	output, _ := card.Transmit(getMemberInfo)

	memberName := ""
	memberId := ""

	if output[1] == 0x80 && output != nil {
		memberId = hex.EncodeToString(output[8:16])
		firstName := parseName(output[16:32])
		lastName := parseName(output[32:48])
		memberName = firstName + " " + lastName
	}
	return MemberInfo{
		memberId: memberId,
		name:     memberName,
	}

}

// Doing some obtuse and definitely Not recommended stuff to squeeze out some more chara length
func parseName(byte []byte) string {
	currentIndex := 0
	var holdover uint8 = 0x00
	var outputBuilder strings.Builder

	for _, o := range byte {
		digitsLeft := 8

		for digitsLeft >= 5 {
			read := o >> (digitsLeft - 5 + currentIndex) & 0x1f
			holdover += read

			if holdover == 31 {
				return outputBuilder.String()
			}

			outputBuilder.WriteRune(rune('A' + holdover))
			holdover = 0

			digitsLeft -= 5 - currentIndex
			currentIndex = 0

		}
		holdover = o << (5 - digitsLeft) & 0x1f
		currentIndex = digitsLeft
	}
	return outputBuilder.String()
}

func encodeName(name string) []byte {
	var output = make([]byte, 16)

	currentIndex := 0
	leftoverIndex := 0
	var leftover = 0x00
	for _, o := range name {
		codeNumber := o - 'A'
		leftover = leftover<<5 + int(codeNumber)
		leftoverIndex += 5

		if leftoverIndex >= 8 {
			output[currentIndex] = byte(leftover >> (leftoverIndex - 8))
			currentIndex++
			leftoverIndex -= 8
		}
	}
	if len(name) <= 24 {
		leftover = leftover<<5 + 31
		leftoverIndex += 5

		if leftoverIndex >= 8 {
			output[currentIndex] = byte(leftover >> (leftoverIndex - 8))
			currentIndex++
			leftoverIndex -= 8
		}
	}

	if leftoverIndex > 0 {
		output[currentIndex] = byte(leftover << (8 - leftoverIndex))
	}

	return output
}

func (cardMan *CardReaderManager) WriteRawPCSC(inputText string) string {

	card, err := cardMan.context.Connect(cardMan.selectedReader, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		return ""
	}

	newContent := strings.ReplaceAll(inputText, " ", "")
	outputBytes, err := hex.DecodeString(newContent)
	if err != nil {
		_ = fmt.Errorf("failed to decode hex: %w", err)
	}
	rsp, _ := card.Transmit(outputBytes)

	var outputBuilder strings.Builder

	for i := 0; i < len(rsp); i++ {
		_, err := fmt.Fprintf(&outputBuilder, "%02x ", rsp[i])
		if err != nil {
			return ""
		}
		if (i+1)%16 == 0 {
			_, err2 := fmt.Fprint(&outputBuilder, "\n")
			if err2 != nil {
				return ""
			}
		}
	}

	err = card.Disconnect(scard.LeaveCard)
	if err != nil {
		return ""
	}
	return outputBuilder.String()
}
