package felica_cards

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
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
	appTab      *container.AppTabs
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

type ScreenInfoMember struct {
	CardID     binding.String
	CardWidget *ReadOnlyEntry
	MemberID   binding.String
	Name       binding.String
}

func NewScreenInfoMember() *ScreenInfoMember {
	cardId := binding.NewString()

	return &ScreenInfoMember{
		CardID:   cardId,
		MemberID: binding.NewString(),
		Name:     binding.NewString(),
	}
}

func GetWindow(pcscInfo *ScreenInfoRawPCSC, memberInfo *ScreenInfoMember) DisplayShow {
	a := app.NewWithID("com.leviwaiu.felicaType")

	a.Preferences().SetBool("HasResponse", false)
	a.Preferences().SetInt("ResponseType", None)
	a.Preferences().SetString("ResponseText", "")
	a.Preferences().SetInt("ResponseInt", 0)

	a.Preferences().SetBool("ReaderBackground", false)

	a.Settings().SetTheme(&generalTheme{})

	w := a.NewWindow("Felica Card Management")

	newDisplayShow := DisplayShow{
		application: a,
		window:      w,
		content:     nil,
	}

	w.SetContent(newDisplayShow.mainBox(pcscInfo, memberInfo))

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
	readString := dis.application.Preferences().String("ResponseText")
	dis.application.Preferences().SetBool("HasResponse", false)
	return responseType, readString
}

func (dis *DisplayShow) mainBox(pcscInfo *ScreenInfoRawPCSC, memberInfo *ScreenInfoMember) *container.AppTabs {

	tabList := container.NewAppTabs(
		container.NewTabItem("Card Info", dis.standardScreen(memberInfo)),
		container.NewTabItem("First Issue", dis.firstIssueScreen()),
		container.NewTabItem("Raw PCSC", dis.rawPCSCScreen(pcscInfo)),
		container.NewTabItem("Settings", dis.settingsScreen()),
	)

	tabList.SetTabLocation(container.TabLocationLeading)
	dis.appTab = tabList

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

func (dis *DisplayShow) standardScreen(memberInfo *ScreenInfoMember) *fyne.Container {

	cardId := NewReadOnlyEntryWithData(memberInfo.CardID)

	memberInfo.CardWidget = cardId
	cardId.ReadOnly()

	cardIdRegex := validation.NewRegexp("\\b[a-f0-9]{16}\\b", "Check for Validity")
	cardId.Validator = cardIdRegex

	memberId := NewReadOnlyEntryWithData(memberInfo.MemberID)
	memberId.ReadOnly()

	memberName := widget.NewEntryWithData(memberInfo.Name)

	points := widget.NewEntry()

	changeButton := widget.NewButton("Change Data", func() {})

	content := container.New(layout.NewFormLayout(),
		widget.NewLabel("Card ID:"),
		cardId,
		widget.NewLabel("Member ID:"),
		memberId,
		widget.NewLabel("Member Name:"),
		memberName,

		widget.NewLabel("Member Since:"),
		widget.NewEntry(),
		widget.NewLabel("Member Expiry:"),
		widget.NewEntry(),

		widget.NewLabel("Current Points"),
		points,

		widget.NewLabel("Entered Programs:"),
		widget.NewTextGrid(),
		layout.NewSpacer(),
		changeButton,
	)
	return content
}

func (dis *DisplayShow) CheckCardBackground() bool {
	if dis.appTab.Selected().Text == "Card Info" {
		return true
	}
	return false
}

func (screenInfo *ScreenInfoMember) UpdateMemberInfo(info MemberInfo) {
	screenInfo.CardID.Set(info.CardID)
	screenInfo.CardWidget.Validate()
	screenInfo.MemberID.Set(info.memberId)
	screenInfo.Name.Set(info.name)

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
