package primitives

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func NewGrid() *tview.Grid {
	d := []int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
	grid := tview.NewGrid().SetRows(d...).SetColumns(d...)
	grid.SetTitleColor(tcell.ColorBlueViolet).
		SetBorderColor(tcell.ColorBlueViolet).
		SetTitle("KYOFI").
		SetBorder(true)

	return grid
}
