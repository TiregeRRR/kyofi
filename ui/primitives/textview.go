package primitives

import "github.com/rivo/tview"

func NewTextView() *tview.TextView {
	view := tview.NewTextView().SetDynamicColors(true)
	view.SetBorder(true)

	return view
}
