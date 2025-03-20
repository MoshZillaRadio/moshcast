package main

import (
	"fmt"
	"moshcast/utils"
)

type Example struct{}

func (e *Example) Process() {
	fmt.Println("🔥 PLUGIN LOAD! 🔥")
}

func Init() utils.Init {
	return &Example{}
}
