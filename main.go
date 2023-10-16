package main

import (
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func NewGrid(row, col []int) *tview.Grid {
	grid := tview.NewGrid().SetRows(row...).SetColumns(col...)
	return grid
}

func main() {
	var dimensions []int
	for i := 0; i < 15; i++ {
		dimensions = append(dimensions, -1)
	}

	grid := NewGrid(dimensions, dimensions)
	grid.SetTitleColor(tcell.ColorBlueViolet).
		SetBorderColor(tcell.ColorBlueViolet).
		SetTitle("KYOFI").
		SetBorder(true)

	listLeft := tview.NewTable().
		SetCell(0, 0, tview.NewTableCell("Name").SetAlign(tview.AlignLeft).SetSelectable(false)).
		SetCell(0, 1, tview.NewTableCell("Size").SetAlign(tview.AlignCenter).SetSelectable(false)).
		SetCell(0, 2, tview.NewTableCell("Permision").SetAlign(tview.AlignRight).SetSelectable(false))
	listRight := tview.NewTable().
		SetCell(0, 0, tview.NewTableCell("Name").SetAlign(tview.AlignLeft).SetSelectable(false)).
		SetCell(0, 1, tview.NewTableCell("Size").SetAlign(tview.AlignCenter).SetSelectable(false)).
		SetCell(0, 2, tview.NewTableCell("Permision").SetAlign(tview.AlignRight).SetSelectable(false))

	itemsLeft, err := getFiles("/home/tireger/Documents")
	if err != nil {
		panic(err)
	}

	itemsRight, err := getFiles("/home/tireger")
	if err != nil {
		panic(err)
	}

	for i := range itemsLeft {
		listLeft.SetCell(i+1, 0, tview.NewTableCell(itemsLeft[i].Name).SetAlign(tview.AlignLeft).SetExpansion(2))
		listLeft.SetCell(i+1, 1, tview.NewTableCell(itemsLeft[i].Size).SetAlign(tview.AlignCenter).SetExpansion(2))
		listLeft.SetCell(i+1, 2, tview.NewTableCell(itemsLeft[i].Permision).SetAlign(tview.AlignRight).SetExpansion(2))
	}

	for i := range itemsRight {
		listRight.SetCell(i+1, 0, tview.NewTableCell(itemsRight[i].Name).SetAlign(tview.AlignLeft).SetExpansion(2))
		listRight.SetCell(i+1, 1, tview.NewTableCell(itemsRight[i].Size).SetAlign(tview.AlignCenter).SetExpansion(2))
		listRight.SetCell(i+1, 2, tview.NewTableCell(itemsRight[i].Permision).SetAlign(tview.AlignRight).SetExpansion(2))
	}

	listLeft.SetSelectable(true, false).SetTitle("root").SetBorder(true)
	listRight.SetSelectable(true, false).SetTitle("ssh").SetBorder(true)

	grid.
		AddItem(listLeft, 0, 0, 15, 8, 0, 80, false).
		AddItem(listRight, 0, 8, 15, 8, 0, 80, false)
	if err := tview.NewApplication().SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

type fileInfo struct {
	Name      string
	Size      string
	Permision string
}

func getFiles(path string) ([]fileInfo, error) {
	f, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	s := make([]fileInfo, len(f))
	for i := range f {
		s[i].Name = f[i].Name()
		inf, err := f[i].Info()
		if err != nil {
			return nil, err
		}
		s[i].Permision = inf.Mode().String()
		s[i].Size = FormatBytes(float64(inf.Size()))
	}

	return s, nil
}
