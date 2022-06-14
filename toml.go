package main

import (
	"log"
	"os"
	"github.com/pelletier/go-toml/v2"
	"regexp"
	"fmt"
)

type gooaTomlPart struct {
	Input 			string
	Output 			string
	Recursive 		bool
	Mapout			bool
	Default_target	string
	Restrict_ext	bool
}

type watchTomlPart struct {
	Watch_interval 	int
	Recompile_on_exit bool
}

type GooaToml struct {
	Gooa gooaTomlPart
	Watch watchTomlPart

	Targets			map[string]string
}

func GenerateTomlConfig(f string) (*GooaToml, bool) {
	t := GooaToml{
		Gooa: gooaTomlPart{},

		Targets: map[string]string{},

	}
	file, err := os.ReadFile(f)

	if err != nil {
		log.Printf("Error opening gooa.toml file (%v)", err)
		return &t, true
	}

	err = toml.Unmarshal(file, &t)

	if err != nil {
		log.Printf("Error parsing gooa.toml file (%v)", err)
		return &t, true
	}
	
	return &t, false
}

func (self *Project) ValidateConfig() bool {
	if !self.Config.Gooa.Mapout && len(self.Config.Targets) == 0 {
		fmt.Printf("[gooa.toml Error] (Needs: [targets]) Must provide atleast one target if you arent directly mapping the folder structure.")
		return true
	}

	if !self.Config.Gooa.Mapout && self.Config.Gooa.Default_target == "" {
		fmt.Printf("[gooa.toml Error] (Needs: [gooa] > default_target) Must provide default target if not directly mapping.")
		return true
	}

	if self.Config.Gooa.Input == "" {
		fmt.Printf("[gooa.toml Error] (Needs: [gooa] > input) Please prove an input directive in your gooa.toml")
		return true
	}

	if self.Config.Gooa.Output == "" {
		fmt.Printf("[gooa.toml Error] (Needs: [gooa] > output) Please prove an output directive in your gooa.toml")
		return true
	}

	if self.Config.Watch.Watch_interval <= 0 {
		self.Config.Watch.Watch_interval = 100
	}

	self.Targets = map[*regexp.Regexp]string{}
	for rgx, out := range self.Config.Targets {
		r, err := regexp.Compile(rgx)

		if err != nil {
			fmt.Printf("[gooa.toml Error] (Invalid Regex) Error Compiling Regex '%v': %v", rgx, err)
			return true
		}

		self.Targets[r] = out
	}

	return false
}