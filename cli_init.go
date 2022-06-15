package main

import (
	"embed"
	"fmt"
	// "os"
	"regexp"
	"strings"
)

//go:embed templates/*
var templates embed.FS

var data map[string]string

func (self *CLI) InitTemplate(dir string, name string, project string) {
	tmpl, err := templates.ReadDir("templates/" + name)

	if err != nil || tmpl == nil {
		tmpl_dirs, err := templates.ReadDir("templates")
		if err != nil {
			fmt.Printf("Failed to read template directory (%v)", err)
			return
		}

		fmt.Printf("Failed to find template '%v', please select one of the following:\n", name)
		for _, v := range tmpl_dirs {
			fmt.Printf(" - %v\n", v.Name())
		}
		return
	}

	data = map[string]string{
		"progname": project,
		"template": name,
		"dir": dir,
	}
	
	self.InitTemplateDirectory(name, "", dir)
}

func (self *CLI) InitTemplateDirectory(dir string, real string, location string) {	
}

var initregex *regexp.Regexp
func (self *CLI) ReplaceTemplateNames(torun string) string {
	if initregex == nil {
		initregex = regexp.MustCompile(`\$\{gooa-([^\s]+)-\}`)
	}

	return initregex.ReplaceAllStringFunc(torun, func(s string) string {
		name := strings.TrimSuffix(strings.TrimPrefix(s, "${gooa-"), "-}")
		
		val, valid := data[name]

		if valid {
			return val
		}
		
		return s
	})
}