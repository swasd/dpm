package main

import (
	"fmt"
	"github.com/Shopify/go-lua"
)

func main() {
	s := lua.NewState()
	s.PushInteger(12)
	s.SetGlobal("cpu")
	s.PushInteger(100)
	s.SetGlobal("i")
	if err := lua.DoString(s, "return (cpu > 10) and (i == 100)"); err != nil {
		panic(err)
	}
	result := s.ToValue(s.Top())
	x := fmt.Sprintf("%v", result)
	fmt.Println(x)
}
