package primitives

import (
	fileinfo "github.com/TiregeRRR/kyofi/file_info"
	"github.com/rivo/tview"
)

func NewTable() *tview.Table {
	table := tview.NewTable().SetSelectable(true, false).SetEvaluateAllRows(true)
	table.SetBorder(true)

	return table
}

func TableHeader(t *tview.Table) {
	t.SetCell(0, 0, tview.NewTableCell("Name").SetAlign(tview.AlignLeft).SetSelectable(false)).
		SetCell(0, 1, tview.NewTableCell("Size").SetAlign(tview.AlignLeft).SetSelectable(false)).
		SetCell(0, 2, tview.NewTableCell("Permision").SetAlign(tview.AlignRight).SetSelectable(false)).
		SetFixed(1, 3)
	t.SetCell(1, 0, tview.NewTableCell("..").SetAlign(tview.AlignLeft).SetSelectable(true))
	t.SetCell(1, 1, tview.NewTableCell("").SetAlign(tview.AlignLeft).SetSelectable(true))
	t.SetCell(1, 2, tview.NewTableCell("").SetAlign(tview.AlignLeft).SetSelectable(true))
}

func UpdateTable(t *tview.Table, fi []fileinfo.FileInfo) {
	t.Clear()
	TableHeader(t)
	for i := range fi {
		t.SetCell(i+2, 0, tview.NewTableCell(fi[i].Name).SetAlign(tview.AlignLeft).SetExpansion(1).SetSelectable(true))
		t.SetCell(i+2, 1, tview.NewTableCell(fi[i].Size).SetAlign(tview.AlignLeft).SetExpansion(1).SetSelectable(true))
		t.SetCell(i+2, 2, tview.NewTableCell(fi[i].Permision).SetAlign(tview.AlignRight).SetExpansion(1).SetSelectable(true))
	}

	t.ScrollToBeginning()
	t.Select(1, 0)
}
