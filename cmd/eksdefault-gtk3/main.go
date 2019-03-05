package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/peterbueschel/eksdefault"
)

const (
	noContext   = "--No Current Context--"
	appName     = "eksdefault-ui"
	columnTitle = "Select EKS Context"
)

var (
	permanent bool
)

type (
	kubeConfig struct {
		curr    string
		currIdx int
		list    []string
		file    *eksdefault.KubeConfig
	}
	chooser struct {
		selection *gtk.TreeSelection
		view      *gtk.TreeView
		store     *gtk.ListStore
		window    *gtk.Window
		box       *gtk.Box
		err       error
		*kubeConfig
	}
)

func (c *chooser) selectionChanged() error {
	model, iter, ok := c.selection.GetSelected()
	if ok {
		tpath, err := model.(*gtk.TreeModel).GetPath(iter)
		if err != nil {
			return err
		}
		iter, err := c.store.GetIter(tpath)
		if err != nil {
			return err
		}
		value, err := c.store.GetValue(iter, 0)
		if err != nil {
			return err
		}
		str, err := value.GetString()
		if err != nil {
			return err
		}
		if str == noContext {
			err = c.kubeConfig.file.UnSetDefault()
		} else {
			err = c.kubeConfig.file.SetContextTo(str)
		}
		if err != nil {
			return err
		}
		c.kubeConfig.curr = str
	}
	return nil
}

func showError(msg string) error {
	c := new(chooser)
	c.setupWindow()
	if c.err != nil {
		return fmt.Errorf("Unable to create window: %s", c.err)
	}
	btn, err := gtk.ButtonNewWithLabel(msg)
	if err != nil {
		return fmt.Errorf("Unable to create button: %s", err)
	}
	_, err = btn.Connect("clicked", func() { c.window.Destroy() })
	if err != nil {
		return fmt.Errorf("Unable to connect button: %s", err)
	}
	c.window.Add(btn)
	c.window.ShowAll()
	if msg == "this is a test" {
		return nil
	}
	gtk.Main()
	return nil
}

func (c *chooser) setupWindow() {
	if c.err != nil {
		return
	}
	if c.window, c.err = gtk.WindowNew(gtk.WINDOW_POPUP); c.err != nil {
		return
	}
	c.window.SetTitle(appName)
	if _, c.err = c.window.Connect("destroy", gtk.MainQuit); c.err != nil {
		return
	}
	c.window.SetPosition(gtk.WIN_POS_MOUSE)
	// need this workaround to support also single click to close the Popup;
	// lost Focus or button-release-event not available for TreeViewNewSelection
	if !permanent {
		_, c.err = c.window.ConnectAfter("button-release-event", func() {
			fmt.Println(c.kubeConfig.curr)
			c.window.Destroy()
		})
	}
}

func (c *chooser) setupRootBox() {
	if c.err != nil {
		return
	}
	c.box, c.err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
}

func (c *chooser) setupSelection() {
	if c.err != nil {
		return
	}
	if c.selection, c.err = c.view.GetSelection(); c.err != nil {
		return
	}
	c.selection.SetMode(gtk.SELECTION_SINGLE)
	path, err := gtk.TreePathNewFromString(fmt.Sprintf("%d", c.kubeConfig.currIdx))
	if err != nil {
		c.err = err
		return
	}
	c.selection.SelectPath(path)
	c.view.RowActivated(path, c.view.GetColumn(0))
	_, c.err = c.selection.Connect("changed", func() error { return c.selectionChanged() })
}

func (c *chooser) setupTreeView() {
	if c.err != nil {
		return
	}
	if c.view, c.err = gtk.TreeViewNewWithModel(c.store); c.err != nil {
		return
	}
	r, err := gtk.CellRendererTextNew()
	if err != nil {
		c.err = err
		return
	}
	column, err := gtk.TreeViewColumnNewWithAttribute(columnTitle, r, "text", 0)
	if err != nil {
		c.err = err
		return
	}
	c.view.AppendColumn(column)
}

func (c *chooser) setupListStore() {
	if c.err != nil {
		return
	}
	if c.store, c.err = gtk.ListStoreNew(glib.TYPE_STRING); c.err != nil {
		return
	}
	for _, i := range c.kubeConfig.list {
		if c.err = c.store.SetValue(c.store.Append(), 0, i); c.err != nil {
			return
		}
	}
}

func initializeChooser(p *kubeConfig) (c *chooser, err error) {
	c = &chooser{kubeConfig: p}
	c.setupListStore()
	c.setupTreeView()
	c.setupSelection()
	c.setupRootBox()
	c.setupWindow()
	if c.err != nil {
		return nil, c.err
	}
	c.box.PackStart(c.view, true, true, 0)
	c.window.Add(c.box)
	return c, nil
}

func fetchContexts() (*kubeConfig, error) {
	p := new(kubeConfig)
	file, err := eksdefault.GetConfigFile()
	if err != nil {
		return nil, err
	}
	p.file = file
	p.list = append(p.file.GetContextNames(), noContext)
	_, p.currIdx, err = p.file.GetContextBy(file.CurrentContext)
	if err != nil || p.currIdx == -2 { // -2 means no default set
		p.curr = noContext
		p.currIdx = len(p.list) - 1 // last item is noContext
	}
	p.curr = file.CurrentContext
	return p, nil
}

func init() {
	flag.BoolVar(&permanent, "permanent", false, "the popup will not be closed after you clicked on a profile")
	flag.Parse()
}

func main() {
	gtk.Init(&os.Args)
	p, err := fetchContexts()
	if err != nil { // only profile related errors
		if e := showError(err.Error()); e != nil {
			log.Println(e)
		}
		log.Fatalln(err)
	}

	c, err := initializeChooser(p)
	if err != nil {
		log.Fatalln(err)
	}
	c.window.ShowAll()
	gtk.Main()
}
