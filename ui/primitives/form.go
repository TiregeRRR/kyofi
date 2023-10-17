package primitives

import "github.com/rivo/tview"

func NewForm() *tview.Form {
	form := tview.NewForm()
	form.SetBorder(true)

	return form
}
