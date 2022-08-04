package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	configFile string
	outputDir  string
)

func init() {
	flag.StringVar(&configFile, "c", "./dir-manager.yaml", "config yaml file")
	flag.StringVar(&outputDir, "o", ".", "output directory")
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()
	b, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	var dm DM
	if err := yaml.Unmarshal(b, &dm); err != nil {
		return err
	}
	for _, v := range dm.Walk(outputDir) {
		if _, err := os.Stat(v.Name); os.IsNotExist(err) {
			var err error
			if v.isFile {
				err = os.WriteFile(v.Name, []byte(""), 0644)
			} else {
				err = os.Mkdir(v.Name, 0755)
			}
			if err != nil {
				return err
			}
		}
	}
	fmt.Println("ok")
	return nil
}

type DM struct {
	GitKeep     bool                   `yaml:"GitKeep"`
	Modules     map[string][]Directory `yaml:"Modules,omitempty"`
	Directories []Directory            `yaml:"Directories,omitempty"`
}

func (d *DM) Walk(basePath string) []*DirFile {
	var fp []*DirFile
	for _, v := range d.Directories {
		fp = append(fp, v.Show(basePath, d.Modules)...)
	}
	return fp
}

type Directory struct {
	Name  string      `yaml:"Name"`
	Child []Directory `yaml:"Child,omitempty"`
	Ref   string      `yaml:"Ref,omitempty"`
}

type DirFile struct {
	isFile bool
	Name   string
}

func (d *Directory) Show(parent string, module map[string][]Directory) []*DirFile {
	var fp []*DirFile
	dn := filepath.Join(parent, d.Name)
	fp = append(fp, &DirFile{Name: dn})
	c := d.Child
	if d.Ref != "" {
		c = module[d.Ref]
	}
	if len(c) == 0 {
		fp = append(fp, &DirFile{isFile: true, Name: filepath.Join(dn, ".gitkeep")})
		return fp
	}
	for _, v := range c {
		fp = append(fp, v.Show(dn, module)...)
	}
	return fp
}
