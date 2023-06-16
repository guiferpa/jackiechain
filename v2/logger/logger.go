package logger

import (
	"log"

	"github.com/fatih/color"
)

func Red(a ...interface{}) {
	log.Println(color.New(color.BgHiRed, color.FgBlack).Sprint(a...))
}

func Yellow(a ...interface{}) {
	log.Println(color.New(color.BgBlack, color.FgHiYellow).Sprint(a...))
}

func Magenta(a ...interface{}) {
	log.Println(color.New(color.BgBlack, color.FgHiMagenta).Sprint(a...))
}
