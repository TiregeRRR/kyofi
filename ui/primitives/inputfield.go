package primitives

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func NewInputField() *tview.InputField {
	input := tview.NewInputField()

	input.SetFieldBackgroundColor(tcell.ColorViolet).SetLabelColor(tcell.ColorViolet)
	return input
}
