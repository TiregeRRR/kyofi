package ui

import (
	fileinfo "github.com/TiregeRRR/kyofi/file_info"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Filer interface {
	Open(string) ([]fileinfo.FileInfo, error)
	Back() ([]fileinfo.FileInfo, error)
}

type App struct {
	grid *tview.Grid

	leftTable  *tview.Table
	rightTable *tview.Table

	leftFiler  Filer
	rightFiler Filer
}

func New(leftFiler, rightFiler Filer) (*App, error) {
	grid := newGrid()
	leftTable := newTable()
	rightTable := newTable()

	leftTable.SetSelectedFunc(func(row, column int) {
		file := leftTable.GetCell(row, column).Text
		var (
			fi  []fileinfo.FileInfo
			err error
		)

		if file == ".." {
			fi, err = leftFiler.Back()
		} else {
			fi, err = leftFiler.Open(file)
		}

		if err != nil {
			panic(err)
		}

		updateTable(leftTable, fi)
	})

	rightTable.SetSelectedFunc(func(row, column int) {
		file := rightTable.GetCell(row, column).Text
		var (
			fi  []fileinfo.FileInfo
			err error
		)

		if file == ".." {
			fi, err = rightFiler.Back()
		} else {
			fi, err = rightFiler.Open(file)
		}

		if err != nil {
			panic(err)
		}

		updateTable(rightTable, fi)
	})

	grid.
		AddItem(leftTable, 0, 0, 15, 8, 0, 80, false).
		AddItem(rightTable, 0, 8, 15, 8, 0, 80, false)

	return &App{
		grid:       grid,
		leftTable:  leftTable,
		rightTable: rightTable,
		leftFiler:  leftFiler,
		rightFiler: rightFiler,
	}, nil
}

func (a *App) Run() error {
	fi, err := a.leftFiler.Open("")
	if err != nil {
		return err
	}

	updateTable(a.leftTable, fi)
	updateTable(a.rightTable, fi)

	if err := tview.NewApplication().SetRoot(a.grid, true).EnableMouse(true).Run(); err != nil {
		return err
	}

	return nil
}

func newGrid() *tview.Grid {
	d := []int{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}
	grid := tview.NewGrid().SetRows(d...).SetColumns(d...)
	grid.SetTitleColor(tcell.ColorBlueViolet).
		SetBorderColor(tcell.ColorBlueViolet).
		SetTitle("KYOFI").
		SetBorder(true)

	return grid
}

func newTable() *tview.Table {
	table := tview.NewTable().SetSelectable(true, false)
	table.SetBorder(true)

	return table
}

func tableHeader(t *tview.Table) {
	t.SetCell(0, 0, tview.NewTableCell("Name").SetAlign(tview.AlignLeft).SetSelectable(false)).
		SetCell(0, 1, tview.NewTableCell("Size").SetAlign(tview.AlignLeft).SetSelectable(false)).
		SetCell(0, 2, tview.NewTableCell("Permision").SetAlign(tview.AlignRight).SetSelectable(false)).
		SetFixed(0, 0).
		SetFixed(0, 1).
		SetFixed(0, 2)
	t.SetCell(1, 0, tview.NewTableCell("..").SetAlign(tview.AlignLeft).SetSelectable(true))
}

func updateTable(t *tview.Table, fi []fileinfo.FileInfo) {
	t.Clear()
	tableHeader(t)
	for i := range fi {
		t.SetCell(i+2, 0, tview.NewTableCell(fi[i].Name).SetAlign(tview.AlignLeft).SetExpansion(2).SetSelectable(true))
		t.SetCell(i+2, 1, tview.NewTableCell(fi[i].Size).SetAlign(tview.AlignLeft).SetExpansion(2).SetSelectable(true))
		t.SetCell(i+2, 2, tview.NewTableCell(fi[i].Permision).SetAlign(tview.AlignRight).SetExpansion(2).SetSelectable(true))
	}

	t.ScrollToBeginning()
	t.Select(1, 0)
}
