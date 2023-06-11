package felica_cards

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type ReadOnlyEntry struct {
	widget.Entry
	readOnly bool
}

func NewReadOnlyEntryWithData(string binding.String) *ReadOnlyEntry {
	entry := &ReadOnlyEntry{
		readOnly: false,
	}
	entry.ExtendBaseWidget(entry)
	entry.Bind(string)

	return entry
}

func (e *ReadOnlyEntry) ReadOnly() {
	e.readOnly = true
}

func (e *ReadOnlyEntry) Writable() {
	e.readOnly = false
}

func (e *ReadOnlyEntry) TypedKey(key *fyne.KeyEvent) {
	if e.readOnly {
		return
	}
	e.Entry.TypedKey(key)
}

func (e *ReadOnlyEntry) TypedRune(r rune) {
	if e.readOnly {
		return
	}
	e.Entry.TypedRune(r)
}

func (e *ReadOnlyEntry) MouseDown(m *desktop.MouseEvent) {
	if e.readOnly {
		return
	}
	e.Entry.MouseDown(m)
}
