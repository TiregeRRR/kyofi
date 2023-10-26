package ui

import (
	"fmt"
	"os"

	fileinfo "github.com/TiregeRRR/kyofi/file_info"
	"github.com/TiregeRRR/kyofi/modules/file"
	"github.com/TiregeRRR/kyofi/ui/primitives"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Copier interface {
	Next() bool
	File() (fileinfo.CopyInfo, error)
}

type Filer interface {
	Open(string) ([]fileinfo.FileInfo, error)
	Back() ([]fileinfo.FileInfo, error)
	Copy(string) error
	PasteReader() (fileinfo.Copier, error)
	Paste(fileinfo.Copier) error
	Delete(string) error
	ProgressLogs() <-chan string
}

type App struct {
	app *tview.Application

	grid *tview.Grid

	leftSide  tview.Primitive
	leftTable *tview.Table

	rightSide  tview.Primitive
	rightTable *tview.Table

	output *tview.TextView

	input *tview.InputField

	leftFiler  Filer
	rightFiler Filer
}

func New() (*App, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	leftFiler, rightFiler := file.New(path), file.New(path)

	grid := primitives.NewGrid()
	leftTable := primitives.NewTable()
	rightTable := primitives.NewTable()

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

		primitives.UpdateTable(leftTable, fi)
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

		primitives.UpdateTable(rightTable, fi)
	})

	output := primitives.NewTextView()

	grid.
		AddItem(leftTable, 0, 0, 15, 8, 0, 80, true).
		AddItem(rightTable, 0, 8, 15, 8, 0, 80, false).
		AddItem(output, 15, 0, 1, 16, 0, 0, false)

	return &App{
		grid:       grid,
		leftTable:  leftTable,
		leftSide:   leftTable,
		rightTable: rightTable,
		rightSide:  rightTable,
		leftFiler:  leftFiler,
		rightFiler: rightFiler,
		output:     output,
		input:      primitives.NewInputField(),
	}, nil
}

func (a *App) handleLogs() {
	go func() {
		for {
			for v := range a.leftFiler.ProgressLogs() {
				a.info(v)
			}
		}
	}()
	go func() {
		for {
			for v := range a.rightFiler.ProgressLogs() {
				a.info(v)
			}
		}
	}()
}

func (a *App) update() error {
	fi, err := a.leftFiler.Open("")
	if err != nil {
		return err
	}

	r, c := a.leftTable.GetSelection()
	primitives.UpdateTable(a.leftTable, fi)
	a.leftTable.Select(r, c)

	fi, err = a.rightFiler.Open("")
	if err != nil {
		return err
	}

	r, c = a.rightTable.GetSelection()
	primitives.UpdateTable(a.rightTable, fi)
	a.rightTable.Select(r, c)

	a.output.ScrollToEnd()

	return nil
}

func (a *App) Run() error {
	fi, err := a.leftFiler.Open("")
	if err != nil {
		return err
	}

	primitives.UpdateTable(a.leftTable, fi)
	primitives.UpdateTable(a.rightTable, fi)

	app := tview.NewApplication()

	a.app = app
	a.app.SetInputCapture(a.handleInput)

	go a.handleLogs()

	if err := a.app.SetRoot(a.grid, true).SetFocus(a.leftTable).EnableMouse(true).Run(); err != nil {
		return err
	}

	return nil
}

func (a *App) handleInput(e *tcell.EventKey) *tcell.EventKey {
	switch e.Key() {
	case tcell.KeyCtrlH:
		a.app.SetFocus(a.leftSide)
	case tcell.KeyLeft:
		a.app.SetFocus(a.leftSide)
	case tcell.KeyCtrlL:
		a.app.SetFocus(a.rightSide)
	case tcell.KeyRight:
		a.app.SetFocus(a.rightSide)
	case tcell.KeyCtrlN:
		if err := a.selectWindow(); err != nil {
			a.err(err)
		}
	}

	foc := a.app.GetFocus()
	switch foc.(type) {
	case *tview.Form:
		return e
	case *tview.InputField:
		return e
	case *tview.Checkbox:
		return e
	}

	switch e.Rune() {
	case 'y':
		if err := a.copy(); err != nil {
			a.err(err)
		}
	case 'p':
		if err := a.paste(); err != nil {
			a.err(err)
		}
	case 'd':
		a.delete()
	}

	if err := a.update(); err != nil {
		a.err(err)
	}

	return e
}

func (a *App) err(err error) {
	a.output.SetText(fmt.Sprintf("%s\n[red]%s[red]", a.output.GetText(false), err.Error()))
}

func (a *App) info(text string) {
	a.output.SetText(fmt.Sprintf("%s\n[white]%s[white]", a.output.GetText(false), text))
}

func (a *App) copy() error {
	var err error

	switch {
	case a.leftTable.HasFocus():
		err = copy(a.leftTable, a.leftFiler)
	case a.rightTable.HasFocus():
		err = copy(a.rightTable, a.rightFiler)
	}

	if err != nil {
		return err
	}

	a.info(fmt.Sprintf("%s copied to buffer", a.getNameInFocus()))

	return nil
}

func (a *App) paste() error {
	var (
		err  error
		name string
	)
	switch {
	case a.leftTable.HasFocus():
		name, err = a._paste(a.leftFiler, a.rightFiler)
	case a.rightTable.HasFocus():
		name, err = a._paste(a.rightFiler, a.leftFiler)

	}

	if err != nil {
		return err
	}

	a.info(fmt.Sprintf("%s pasted", name))

	return nil
}

func (a *App) delete() {
	var err error

	name := a.getNameInFocus()

	deleteFunc := func() error {
		switch {
		case a.leftTable.HasFocus():
			err = delete(a.leftTable, a.leftFiler)
		case a.rightTable.HasFocus():
			err = delete(a.rightTable, a.rightFiler)
		default:
			return nil
		}

		if err != nil {
			return err
		}

		a.info(fmt.Sprintf("%s deleted", name))

		return nil
	}

	lastFocus := a.app.GetFocus()

	a.input.SetLabel(fmt.Sprintf("delete %s?: y/n  ", name)).SetDoneFunc(func(key tcell.Key) {
		defer a.grid.RemoveItem(a.input)
		defer a.app.SetFocus(lastFocus)

		if key == tcell.KeyEnter {
			switch a.input.GetText() {
			case "y", "yes":
				a.app.SetFocus(lastFocus)
				if err := deleteFunc(); err != nil {
					a.err(err)
				}
			}
		}

		if err := a.update(); err != nil {
			a.err(err)
		}
	})

	a.input.SetText("")
	a.grid.AddItem(a.input, 15, 0, 1, 16, 0, 0, false)
	a.app.SetFocus(a.input)
}

func (a *App) swapContextsWithFiler(cur tview.Primitive, tar tview.Primitive, fi Filer) {
	switch {
	case a.leftSide == cur:
		a.leftFiler = fi
	case a.rightSide == cur:
		a.rightFiler = fi
	}

	a.swapContexts(cur, tar)
}

func (a *App) swapContexts(cur tview.Primitive, tar tview.Primitive) {
	switch {
	case a.leftSide == cur:
		a.grid.RemoveItem(cur)
		a.grid.AddItem(tar, 0, 0, 15, 8, 0, 80, true)
		a.leftSide = tar
		a.app.SetFocus(a.leftSide)
	case a.rightSide == cur:
		a.grid.RemoveItem(cur)
		a.grid.AddItem(tar, 0, 0, 15, 8, 0, 80, true)
		a.rightSide = tar
		a.app.SetFocus(a.rightSide)
	}
}

func (a *App) getNameInFocus() string {
	switch {
	case a.leftTable.HasFocus():
		r, _ := a.leftTable.GetSelection()
		return a.leftTable.GetCell(r, 0).Text
	case a.rightTable.HasFocus():
		r, _ := a.rightTable.GetSelection()
		return a.rightTable.GetCell(r, 0).Text
	}

	return ""
}

func copy(table *tview.Table, filer Filer) error {
	r, _ := table.GetSelection()
	p := table.GetCell(r, 0).Text
	if p == ".." {
		return nil
	}

	if err := filer.Copy(p); err != nil {
		return err
	}

	return nil
}

func (a *App) _paste(dst Filer, src Filer) (string, error) {
	cop, err := src.PasteReader()
	if err != nil {
		return "", err
	}

	go func() {
		if err := dst.Paste(cop); err != nil {
			a.err(err)
		}
	}()

	return "", nil
}

func delete(table *tview.Table, filer Filer) error {
	r, _ := table.GetSelection()
	p := table.GetCell(r, 0).Text
	if p == ".." {
		return nil
	}

	if err := filer.Delete(p); err != nil {
		return err
	}

	return nil
}
