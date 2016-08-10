package main

import (
	"flag"
	"io/ioutil"

	"github.com/labstack/echo/engine/standard"
	"github.com/rpoletaev/testauth/auth"
	"gopkg.in/yaml.v2"
)

func main() {
	init := flag.Bool("init", false, "Create tables and common entities")
	flag.Parse()
	config := &auth.Config{}
	configBytes, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(configBytes, config)
	if err != nil {
		panic(err)
	}

	api, erro := auth.NewAPI(*config, *init)
	if erro != nil {
		panic(erro)
	}

	api.Run(standard.New(config.Port))
}
