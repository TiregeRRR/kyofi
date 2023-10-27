package ui

import (
	"os"
	"sync/atomic"

	"github.com/TiregeRRR/kyofi/modules/file"
	"github.com/TiregeRRR/kyofi/modules/minio"
	"github.com/TiregeRRR/kyofi/modules/ssh"
	"github.com/TiregeRRR/kyofi/ui/primitives"
	"github.com/rivo/tview"
)

func (a *App) selectWindow() error {
	var err error

	switch {
	case a.leftSide.HasFocus():
		err = a.drawSelectWindow(a.leftSide, true)
	case a.rightSide.HasFocus():
		err = a.drawSelectWindow(a.rightSide, false)
	}

	if err != nil {
		return err
	}

	return nil
}

func (a *App) drawSelectWindow(p tview.Primitive, left bool) error {
	slc := a.selectTable()
	a.swapContexts(p, slc)

	slc.SetSelectedFunc(func(row, column int) {
		name := slc.GetCell(row, column).Text
		switch name {
		case "File":
			path, err := os.Getwd()
			if err != nil {
				a.err(err)

				return
			}
			if left {
				a.leftFiler.Close()
				a.leftFiler = file.New(path)
			} else {
				a.rightFiler.Close()
				a.rightFiler = file.New(path)
			}
		case "Minio":
			form := a.minioSelectForm(p, left)
			a.swapContexts(slc, form)

			return
		case "SSH":
			form := a.sshSelectForm(p, left)
			a.swapContexts(slc, form)
		}

		a.swapContexts(slc, p)
		if err := a.update(); err != nil {
			a.err(err)
		}
	})

	return nil
}

func (a *App) selectTable() *tview.Table {
	table := primitives.NewTable().
		SetCell(0, 0, tview.NewTableCell("File")).
		SetCell(1, 0, tview.NewTableCell("Minio")).
		SetCell(2, 0, tview.NewTableCell("SSH"))

	return table
}

func (a *App) minioSelectForm(p tview.Primitive, left bool) *tview.Form {
	form := primitives.NewForm()

	form.SetBorder(true).SetTitle("Minio")

	exit := func() {
		a.swapContexts(form, p)
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

		a.swapContextsWithFiler(form, p, m)

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

func (a *App) sshSelectForm(p tview.Primitive, left bool) *tview.Form {
	form := primitives.NewForm()

	form.SetBorder(true).SetTitle("SSH")

	exit := func() {
		a.swapContexts(form, p)
	}

	var conn atomic.Bool

	connect := func() {
		if conn.Load() {
			return
		}
		conn.Store(true)

		m, err := ssh.New(ssh.Opts{
			Addr:     form.GetFormItem(0).(*tview.InputField).GetText(),
			User:     form.GetFormItem(1).(*tview.InputField).GetText(),
			Password: form.GetFormItem(2).(*tview.InputField).GetText(),
		})
		if err != nil {
			a.err(err)
			exit()
			return
		}

		a.swapContextsWithFiler(form, p, m)

		if err := a.update(); err != nil {
			a.err(err)
		}
	}

	form.
		AddInputField("Address", "", 50, nil, nil).
		AddInputField("User", "", 50, nil, nil).
		AddPasswordField("Password", "", 50, '*', nil).
		AddButton("Connect", connect).
		AddButton("Exit", exit)

	return form
}
