package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/radovskyb/watcher"
)

func wprint(f string) {
	fmt.Printf("]  %v\n", f)
}

func (self *Project) Watch(dir string) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		if self.Config.Watch.Recompile_on_exit {
			fmt.Print("]---------------------------------------\n")
			wprint("Running Exit Compile")

			cerr := self.Compile(dir)

			if cerr != nil {
				wprint("Exit Compile Failed :(, You might have an issue")
				return
			}
			wprint("Exit Compile Succeeded :)")
			fmt.Print("]---------------------------------------\n")
		} else {
			fmt.Print("]---------------------------------------\n")
			wprint("]  Exiting\n")
			fmt.Print("]---------------------------------------\n")
		}

		os.Exit(0)
	}()


	self.Metadata = true
	fmt.Print("\n]---------------------------------------\n")
	wprint("GooaC Watcher  (Press Ctrl+C To Exit)")
	if !self.Config.Gooa.Mapout {
		wprint("Remember to recompile after youre done developing to remove useless metadata!")
	}
	wprint("Running Initial Compile")

	cerr := self.Compile(dir)

	if cerr != nil {
		wprint("]  Initial Compile Failed, Exiting Watcher\n")
		fmt.Printf("]---------------------------------------\n")
		return
	}

	w := watcher.New()
	w.FilterOps(
		watcher.Write,
		watcher.Move,
		watcher.Create,
		watcher.Remove,
		watcher.Rename,
	)

	go func() {
		self.WatchRoutine(w)
	}()

	if self.Config.Gooa.Recursive {
		err := w.AddRecursive(dir + self.Config.Gooa.Input)

		if err != nil {
			wprint("Error Adding Recursive Watcher (" + err.Error() + ")")
			return
		}
	} else {
		err := w.Add(dir + self.Config.Gooa.Input)

		if err != nil {
			wprint("Error Adding Watcher (" + err.Error() + ")")
			return
		}
	}

	err := w.Start(time.Millisecond * time.Duration(self.Config.Watch.Watch_interval))

	if err != nil {
		wprint("Error Starting Watcher (" + err.Error() + ")")
	}
}

func (self *Project) WatchRoutine(w *watcher.Watcher) {
	for {
		select {
		case event := <-w.Event:
			switch event.Op {
			case watcher.Write: 	self.WatchHandleWrite(w, event)
			case watcher.Create: 	self.WatchHandleCreate(w, event)
			case watcher.Remove: 	self.WatchHandleRemove(w, event)
			case watcher.Move: 		self.WatchHandleMove(w, event)
			case watcher.Rename: 	self.WatchHandleRename(w, event)
			}
		
		case err := <-w.Error:
			switch err {
			case watcher.ErrWatchedFileDeleted:
			default:
				wprint("Error While Watching (" + err.Error() + ")")
			}
		
		case <-w.Closed:
			return
		}
	}
}

