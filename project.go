package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gooac/gooac"
)

type Project struct {
	Root 		string
	Config 		*GooaToml
	Compiler	*gooa.Gooa
	Targets 	map[*regexp.Regexp]string
	
	Metadata 	bool
	RegexCache	map[string]*regexp.Regexp
}

func NewProject(dir string) (Project, bool) {
	pj := Project{}
	toml, err := GenerateTomlConfig(dir + "gooa.toml")

	if err {
		return pj, true
	}

	pj.Root = filepath.Clean(dir)
	pj.Compiler = gooa.NewGooa()
	pj.Config = toml
	pj.Metadata = false
	pj.RegexCache = map[string]*regexp.Regexp{}

	return pj, pj.ValidateConfig()
}

func (self *Project) Compile(dir string) error {
	os.RemoveAll(dir + self.Config.Gooa.Output)
	os.MkdirAll(dir + self.Config.Gooa.Output, 0777)	
	cerr := self.CompileDirectory(dir + self.Config.Gooa.Input, self.Config.Gooa.Recursive, "", dir + filepath.Clean(self.Config.Gooa.Output))

	if cerr != nil {
		return cerr
	}

	return nil
}

func (self *Project) Where(strippedname string) string {
	outfile := self.Config.Gooa.Default_target

	for rgx, fl := range self.Targets {
		if rgx.Match([]byte(strippedname)) {
			outfile = fl
			break
		}
	}

	return outfile
}

func (self *Project) CompileDirectory(dir string, recur bool, relative string, outdir string) error {
	lst, err := ioutil.ReadDir(dir)

	if err != nil {
		return err
	}

	for _, v := range lst {
		if !v.IsDir() {
			ext := filepath.Ext(v.Name())

			if self.Config.Gooa.Mapout {
				if ext != ".lua" && ext != ".gooa" && self.Config.Gooa.Restrict_ext {
					return nil
				}

				self.CompileFile(dir, v.Name(), outdir + relative + "/" + v.Name(), false)
				self.Compiler.Err.Reset()
			} else {
				strippedname := strings.TrimSuffix(v.Name(), ext)
				
				outfile := self.Where(strippedname)
				if self.Config.Gooa.Restrict_ext {
					if ext != ".lua" && ext != ".gooa" {
						return nil
					}
				}

				self.CompileFile(dir, v.Name(), outdir + "/" + outfile, false)
				self.Compiler.Err.Reset()
			}
		} else if recur {
			if self.Config.Gooa.Mapout {
				cerr := os.Mkdir(outdir + relative + "/" + v.Name(), 0777)
				
				if cerr != nil {
					fmt.Printf("Error Creating Directory %v: %v\n", outdir + relative + "/" + v.Name(), cerr)
				}
			}

			if comperr := self.CompileDirectory(dir + v.Name() + "/", true, relative + "/" + v.Name(), outdir); comperr != nil {
				return comperr
			}
		}
	}

	return nil
}

func (self *Project) CompileFile(dir string, name string, toout string, erase bool) error {
	dat, err := self.Compiler.CompileFile(dir + name)

	if err != nil {
		fmt.Printf("Failed to compile %v (%v)\n", dir + name, err)
		return err
	}
	
	if erase { 
		os.Remove(toout)
	}

	f, werr := os.OpenFile(toout, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0757)
	
	if werr != nil {
		fmt.Printf("Error Opening %v: %v\n", toout, werr)
		return werr
	}
	
	defer f.Close()
	
	if self.Metadata && !self.Config.Gooa.Mapout {
		n := strings.ReplaceAll(dir + name, "\\", "/")
		f.WriteString("--[[" + n + ">]]")
		_, werr = f.WriteString(dat)
		f.WriteString("--[[<" + n + "]]")
	} else {
		_, werr = f.WriteString(dat + "\n")
	}

	if werr != nil {
		fmt.Printf("Error Writing to %v: %v\n", toout, werr)
		return werr
	}
	
	return nil
}