package ui

import (
	"os"
	"sync/atomic"

	"github.com/TiregeRRR/kyofi/modules/file"
	"github.com/TiregeRRR/kyofi/modules/minio"
	"github.com/TiregeRRR/kyofi/ui/primitives"
	"github.com/rivo/tview"
)

func (a *App) selectWindow() error {
	var err error

	switch {
	case a.leftTable.HasFocus():
		err = a.drawSelectWindow(a.leftTable, true)
	case a.rightTable.HasFocus():
		err = a.drawSelectWindow(a.rightTable, false)
	}

	if err != nil {
		return err
	}

	return nil
}

func (a *App) drawSelectWindow(p tview.Primitive, left bool) error {
	slc := a.selectTable()
	if left {
		a.grid.RemoveItem(p).AddItem(slc, 0, 0, 15, 8, 0, 80, true)
	} else {
		a.grid.RemoveItem(p).AddItem(slc, 0, 8, 15, 8, 0, 80, false)
	}
	a.app.SetFocus(slc)

	slc.SetSelectedFunc(func(row, column int) {
		name := slc.GetCell(row, column).Text
		switch name {
		case "File":
			path, err := os.Getwd()
			if err != nil {
				a.err(err)
			}
			if left {
				a.leftFiler = file.New(path)
				a.leftTable.SetBorder(true).SetTitle("File")
			} else {
				a.rightFiler = file.New(path)
				a.rightTable.SetBorder(true).SetTitle("File")
			}
		case "Minio":
			form := a.minioSelectForm(left)
			a.grid.RemoveItem(slc)
			if left {
				a.grid.AddItem(form, 0, 0, 15, 8, 0, 80, true)
				a.leftSide = form
			} else {
				a.grid.AddItem(form, 0, 8, 15, 8, 0, 80, false)
				a.rightSide = form
			}

			a.app.SetFocus(form)

			return
		}

		a.grid.RemoveItem(slc)
		a.grid.AddItem(p, 0, 0, 15, 8, 0, 80, true)
		a.app.SetFocus(p)
		if err := a.update(); err != nil {
			a.err(err)
		}
	})

	return nil
}

func (a *App) selectTable() *tview.Table {
	table := primitives.NewTable().
		SetCell(0, 0, tview.NewTableCell("File").SetAlign(tview.AlignCenter)).
		SetCell(1, 0, tview.NewTableCell("Minio").SetAlign(tview.AlignCenter))

	return table
}

func (a *App) minioSelectForm(left bool) *tview.Form {
	form := primitives.NewForm()

	exit := func() {
		a.grid.RemoveItem(form)

		if left {
			a.grid.AddItem(a.leftTable, 0, 0, 15, 8, 0, 80, true)
			a.leftSide = a.leftTable
			a.app.SetFocus(a.leftSide)
		} else {
			a.grid.AddItem(a.rightTable, 0, 8, 15, 8, 0, 80, false)
			a.rightSide = a.rightTable
			a.app.SetFocus(a.rightSide)
		}
	}

	var conn atomic.Bool

	connect := func() {
		if conn.Load() {
			return
		}
		conn.Store(true)

		m, err := minio.New(minio.Opts{
			Endpoint:        form.GetFormItem(0).(*tview.InputField).GetText(),
			AccessKeyID:     form.GetFormItem(1).(*tview.InputField).GetText(),
			SecretAccessKey: form.GetFormItem(2).(*tview.InputField).GetText(),
			UseSSL:          form.GetFormItem(3).(*tview.Checkbox).IsChecked(),
			SkipVerify:      form.GetFormItem(4).(*tview.Checkbox).IsChecked(),
		})
		if err != nil {
			a.err(err)
			exit()
			return
		}

		a.grid.RemoveItem(form)

		if left {
			a.grid.AddItem(a.leftTable, 0, 0, 15, 8, 0, 80, true)
			a.leftSide = a.leftTable
			a.leftFiler = m
		} else {
			a.grid.AddItem(a.rightTable, 0, 8, 15, 8, 0, 80, false)
			a.rightSide = a.rightTable
			a.rightFiler = m
		}

		if err := a.update(); err != nil {
			a.err(err)
		}
	}

	form.
		AddInputField("Endpoint", "", 50, nil, nil).
		AddInputField("Access Key ID", "", 50, nil, nil).
		AddInputField("Secret Access Key", "", 50, nil, nil).
		AddCheckbox("Use SSL", false, nil).
		AddCheckbox("Insecure", false, nil).
		AddButton("Connect", connect).
		AddButton("Exit", exit)

	return form
}
