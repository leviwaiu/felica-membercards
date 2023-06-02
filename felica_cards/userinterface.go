package felica_cards

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"time"
)

const (
	None            = 0
	NewReaderSelect = 1001
	SubmitPCSCCode  = 1002
)

type DisplayShow struct {
	application fyne.App
	window      fyne.Window
	content     *fyne.Container
}

type ScreenInfoRawPCSC struct {
	ReaderList     []string
	ReaderSelect   *widget.Select
	PcscInput      binding.String
	textGridOutput *widget.TextGrid
}

func NewScreenInfoRawPCSC() *ScreenInfoRawPCSC {
	return &ScreenInfoRawPCSC{
		ReaderList: []string{},
		PcscInput:  binding.NewString(),
	}
}

func GetWindow(pcscInfo *ScreenInfoRawPCSC) DisplayShow {
	a := app.NewWithID("com.leviwaiu.felicaType")

	a.Preferences().SetBool("HasResponse", false)
	a.Preferences().SetInt("ResponseType", None)
	a.Preferences().SetString("ResponseText", "")

	w := a.NewWindow("Felica Card Management")

	newDisplayShow := DisplayShow{
		application: a,
		window:      w,
		content:     nil,
	}

	w.SetContent(newDisplayShow.mainBox(pcscInfo))

	return newDisplayShow
}

func (dis *DisplayShow) StartWindow() {
	dis.window.ShowAndRun()
}

func (dis *DisplayShow) GetEvent() (int, string) {
	for dis.application.Preferences().Bool("HasResponse") == false {
		time.Sleep(100 * time.Millisecond)
	}
	responseType := dis.application.Preferences().Int("ResponseType")
	readerString := dis.application.Preferences().String("ResponseText")
	dis.application.Preferences().SetBool("HasResponse", false)
	return responseType, readerString
}

func (dis *DisplayShow) mainBox(pcscInfo *ScreenInfoRawPCSC) *container.AppTabs {

	tabList := container.NewAppTabs(
		container.NewTabItem("Card Info", dis.standardScreen()),
		container.NewTabItem("First Issue", dis.firstIssueScreen()),
		container.NewTabItem("Raw PCSC", dis.rawPCSCScreen(pcscInfo)),
		container.NewTabItem("Settings", dis.settingsScreen()),
	)

	tabList.SetTabLocation(container.TabLocationLeading)

	return tabList
}

func (dis *DisplayShow) rawPCSCScreen(pcscInfo *ScreenInfoRawPCSC) *fyne.Container {

	pcscInfo.ReaderSelect = widget.NewSelect(
		pcscInfo.ReaderList, func(s string) {
			dis.application.Preferences().SetString("ResponseText", s)
			dis.application.Preferences().SetInt("ResponseType", NewReaderSelect)
			dis.application.Preferences().SetBool("HasResponse", true)
		})

	textGrid := widget.NewTextGrid()

	pcscInfo.textGridOutput = textGrid
	textGrid.SetText("")

	pcscInputSpace := widget.NewEntryWithData(pcscInfo.PcscInput)

	executeFunc := func() {
		dis.application.Preferences().SetInt("ResponseType", SubmitPCSCCode)
		dis.application.Preferences().SetBool("HasResponse", true)
	}

	enterButton := widget.NewButton("Submit Instruction", executeFunc)

	content := container.New(layout.NewVBoxLayout(),
		widget.NewLabel("Selected Reader:"),
		pcscInfo.ReaderSelect,
		widget.NewLabel("Input your PCSC Binary Code here:"),
		pcscInputSpace,
		widget.NewLabel("Output"),
		textGrid,
		layout.NewSpacer(),
		enterButton,
	)

	return content
}

func (screenInfo *ScreenInfoRawPCSC) AutoChooseReader(reader string) {
	screenInfo.ReaderSelect.SetSelected(reader)
}

func (screenInfo *ScreenInfoRawPCSC) UpdateOutput(s string) {
	screenInfo.textGridOutput.SetText(s)
}

func (dis *DisplayShow) firstIssueScreen() *fyne.Container {

	idInput := widget.NewEntry()

	ckInput := widget.NewEntry()

	ndefEnable := widget.NewCheck("Enable Ndef", BoogeyFunc)

	useVerification := widget.NewCheck("Enable CKV and CK block writing with MAC", BoogeyFunc)

	confirmButton := widget.NewButton("Confirm First Issue", nil)

	content := container.New(layout.NewVBoxLayout(),
		widget.NewLabel("ID Layout Input"),
		idInput,
		widget.NewLabel("Card Key"),
		ckInput,
		ndefEnable,
		useVerification,
		layout.NewSpacer(),
		confirmButton,
	)
	return content
}

func (dis *DisplayShow) standardScreen() *fyne.Container {

	cardBinding := binding.NewString()
	cardId := widget.NewEntryWithData(cardBinding)

	idBinding := binding.NewString()
	memberId := widget.NewEntryWithData(idBinding)
	memberId.Disable()

	nameBinding := binding.NewString()
	memberName := widget.NewEntryWithData(nameBinding)

	content := container.New(layout.NewFormLayout(),
		widget.NewLabel("Card ID:"),
		cardId,
		widget.NewLabel("Member ID:"),
		memberId,
		widget.NewLabel("Member Name:"),
		memberName,
		widget.NewLabel("Member Since:"),
		widget.NewLabel("Member Expiry:"),
	)
	return content
}

func (dis *DisplayShow) settingsScreen() *fyne.Container {

	pcscSelect := widget.NewSelect([]string{}, func(x string) {

	})

	content := container.New(layout.NewFormLayout(),
		widget.NewLabel("Selected Reader:"),
		pcscSelect,
	)
	return content
}
