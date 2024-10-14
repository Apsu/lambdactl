package main

import (
	"embed"

	"lambdactl/cmd"
)

//go:embed deploy/*
var lambdaFS embed.FS

func main() {
	cmd.Execute(lambdaFS)
}
