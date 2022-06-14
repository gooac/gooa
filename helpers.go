package main

import (
	"errors"
	"path/filepath"
	"strings"
)

type ProjectPathInfo struct {
	RealPath string
	Split []string

	FoundIn string
	Relative string
	RelativeInclusive string
}

func (self Project) AbsolutePathToProjectInfo(p string) (error, ProjectPathInfo) {
	info := ProjectPathInfo{}

	path := filepath.Clean(p)
	split := strings.SplitN(path, self.Root, 2)
	
	info.RealPath = path
	info.Split = split

	if len(split) == 1 {
		return errors.New("Failed to find root in path"), info
	}

	incl := filepath.Clean(split[1])
	info.RelativeInclusive = incl

	in := filepath.Clean(self.Config.Gooa.Input)
	insplit := strings.SplitN(incl, in, 2)
	
	if len(insplit) == 2 {
		info.FoundIn = in
		info.Relative = strings.TrimLeft(insplit[1], "/\\")
	}

	out := filepath.Clean(self.Config.Gooa.Output)
	outsplit := strings.SplitN(incl, in, 2)
	
	if len(outsplit) == 1 {
		return errors.New("Failed to find prefix in path"), info
	}

	info.FoundIn = out
	info.Relative = strings.TrimLeft(outsplit[1], "/\\")

	return nil, info
}

func (self Project) IsSameRoot(l string, r string) (bool, ProjectPathInfo, ProjectPathInfo) {
	lerr, lpath := self.AbsolutePathToProjectInfo(l)
	rerr, rpath := self.AbsolutePathToProjectInfo(r)

	if (lerr != nil || rerr != nil) || (lpath.Split[0] != rpath.Split[0]) || (lpath.FoundIn != rpath.FoundIn) {
		return false, lpath, rpath
	}

	return true, lpath, rpath
}