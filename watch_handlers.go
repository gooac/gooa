package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/radovskyb/watcher"
)

func (self *Project) WatchHandleWrite(w *watcher.Watcher, e watcher.Event) {	
	if e.IsDir() {
		return
	}

	err, info := self.AbsolutePathToProjectInfo(e.Path)
	
	if err != nil {
		wprint("Error Parsing Path Information (" + err.Error() +")")
		return
	}

	if self.Config.Gooa.Mapout {
		self.Compiler.Err.Reset()
		cerr := self.CompileFile(self.Root + info.RelativeInclusive, "",
		self.Root + "/" + self.Config.Gooa.Output + "/" + info.Relative, true)

		if cerr != nil {
			wprint("Failed to compile '" + info.RelativeInclusive + "'")
			return
		}

		wprint("Successfully Recompiled '" + info.RelativeInclusive + "'")
	} else {
		dir := strings.ReplaceAll(self.Root + info.RelativeInclusive, "\\", "/")

		self.Compiler.Err.Reset()
		out, err := self.Compiler.CompileFile(self.Root + info.RelativeInclusive)

		if err != nil {
			fmt.Printf("Failed to compile %v (%v)\n", self.Root + info.RelativeInclusive, err)
			return
		}

		rgx, val := self.RegexCache[dir]

		if !val {
			rgx, _ = regexp.Compile(`\-\-\[\[` + dir + `>\]\](.*)--\[\[<` + dir + `\]\]`)
			self.RegexCache[dir] = rgx
		}

		where := filepath.Clean(self.Root + "/" + self.Config.Gooa.Output + "/" + self.Where(dir))
		f, err := os.ReadFile(where)

		if err != nil {
			fmt.Printf("Failed to open compiled file '%v' (due to '%v'): %v\n", where, dir, err)
			return
		}

		rep := rgx.ReplaceAllLiteralString(string(f), fmt.Sprintf("--[[%v>]]%v--[[<%v]]", dir, out, dir))

		os.Remove(where)
		os.WriteFile(where, []byte(rep), 0757)
	
		wprint("Successfully Recompiled '" + dir + "' and injected into compiled file '" + where + "'")
	}
}

func (self *Project) WatchHandleCreate(w *watcher.Watcher, e watcher.Event) {	
	err, info := self.AbsolutePathToProjectInfo(e.Path)

	if err != nil {
		wprint("Error Parsing Path Information (" + err.Error() +")")
		return
	}

	if e.IsDir() {
		w.AddRecursive(e.Path)

		dir := self.Root + "/" + self.Config.Gooa.Output + info.Relative
		merr := os.MkdirAll(dir, 0757)

		if merr != nil {
			wprint("Failed to create directory '" + dir + "' (" + merr.Error() + ")")
			return
		}

		wprint("Created directory '" + dir + "'")
		return
	}
	
	if self.Config.Gooa.Mapout {
		self.Compiler.Err.Reset()
		cerr := self.CompileFile(self.Root + info.RelativeInclusive, "",
		self.Root + "/" + self.Config.Gooa.Output + "/" + info.Relative, true)
	
		if cerr != nil {
			wprint("Failed to compile '" + info.RelativeInclusive + "'")
			return
		}
	
		wprint("Successfully compiled '" + info.RelativeInclusive + "'")
	} else {
		dir := strings.ReplaceAll(self.Root + info.RelativeInclusive, "\\", "/")

		self.Compiler.Err.Reset()
		out, err := self.Compiler.CompileFile(self.Root + info.RelativeInclusive)

		if err != nil {
			fmt.Printf("Failed to compile %v (%v)\n", self.Root + info.RelativeInclusive, err)
			return
		}

		rgx, val := self.RegexCache[dir]

		if !val {
			rgx, _ = regexp.Compile(`\-\-\[\[` + dir + `>\]\](.*)--\[\[<` + dir + `\]\]`)
			self.RegexCache[dir] = rgx
		}

		where := filepath.Clean(self.Root + "/" + self.Config.Gooa.Output + "/" + self.Where(dir))
		f, err := os.OpenFile(where, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0757)

		if err != nil {
			fmt.Printf("Failed to open compiled file '%v' (due to '%v'): %v\n", where, dir, err)
			return
		}

		f.Write([]byte(fmt.Sprintf("--[[%v>]]%v--[[<%v]]", dir, out, dir)))
		f.Close()

		wprint("Successfully compiled '" + dir + "' and injected into compiled file '" + where + "'")
	}
}

func (self *Project) WatchHandleRemove(w *watcher.Watcher, e watcher.Event) {
	err, info := self.AbsolutePathToProjectInfo(e.Path)

	if err != nil {
		wprint("Error Parsing Path Information (" + err.Error() +")")
		return
	}

	dir := self.Root + "/" + self.Config.Gooa.Output + info.Relative

	if e.IsDir() {
		os.RemoveAll(dir)
		wprint("Removed directory '" + dir + "'")
		return
	}

	if self.Config.Gooa.Mapout {
		os.RemoveAll(dir)
		wprint("Removed File '" + dir + "'")
		return
	}

	dir = strings.ReplaceAll(self.Root + info.RelativeInclusive, "\\", "/")

	rgx, val := self.RegexCache[dir]

	if !val {
		rgx, _ = regexp.Compile(`\-\-\[\[` + dir + `>\]\](.*)--\[\[<` + dir + `\]\]`)
		self.RegexCache[dir] = rgx
	}

	where := filepath.Clean(self.Root + "/" + self.Config.Gooa.Output + "/" + self.Where(dir))
	f, err := os.ReadFile(where)

	if err != nil {
		fmt.Printf("Failed to open compiled file '%v' (due to '%v'): %v\n", where, dir, err)
		return
	}

	rep := rgx.ReplaceAllLiteralString(string(f), "")

	os.Remove(where)
	os.WriteFile(where, []byte(rep), 0757)

	wprint("Successfully removed '" + dir + "' from compiled file '" + where + "'")
}

// I've never actually seen this event get triggered
// so..... heres this?
// if you do encounter this please make an issue
func (self *Project) WatchHandleMove(w *watcher.Watcher, e watcher.Event) {
	same, _, _ := self.IsSameRoot(e.OldPath, e.Path)

	if !same {
		if e.IsDir() {
			w.RemoveRecursive(e.Path)
		}

		wprint("Lost sight of '" + e.OldPath + "', moved to '" + e.Path + "'")
	}

	wprint("Moved '" + e.OldPath + "' to '" + e.Path + "'")
}

func (self *Project) WatchHandleRename(w *watcher.Watcher, e watcher.Event) {
	_, old, new := self.IsSameRoot(e.OldPath, e.Path)

	if self.Config.Gooa.Mapout {
		os.Rename(self.Root + "/" + self.Config.Gooa.Output + "/" + old.Relative, self.Root + "/" + self.Config.Gooa.Output + "/" + new.Relative)
	} else {
		err, newinfo := self.AbsolutePathToProjectInfo(e.Path)
		if err != nil {
			wprint("Error Parsing Path Information (" + err.Error() +")")
			return
		}

		err, info := self.AbsolutePathToProjectInfo(e.OldPath)
		if err != nil {
			wprint("Error Parsing Path Information (" + err.Error() +")")
			return
		}
	
		dir := strings.ReplaceAll(self.Root + info.RelativeInclusive, "\\", "/")
		newdir := strings.ReplaceAll(self.Root + newinfo.RelativeInclusive, "\\", "/")

		rgx, val := self.RegexCache[dir]
		if !val {
			rgx, _ = regexp.Compile(`\-\-\[\[` + dir + `>\]\](.*)--\[\[<` + dir + `\]\]`)
			self.RegexCache[dir] = rgx
		}

		where := filepath.Clean(self.Root + "/" + self.Config.Gooa.Output + "/" + self.Where(dir))
		f, err := os.ReadFile(where)

		if err != nil {
			fmt.Printf("Failed to open compiled file '%v' (due to '%v'): %v\n", where, dir, err)
			return
		}

		rep := rgx.ReplaceAllString(string(f), fmt.Sprintf("--[[%v>]]$1--[[<%v]]", newdir, newdir))

		os.Remove(where)
		os.WriteFile(where, []byte(rep), 0757)
	
		wprint("Successfully Renamed '" + dir + "' in '" + where + "'")
	}
}