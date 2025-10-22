package logging

import (
	"fmt"
	"time"
)

func Info(format string, a ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("[%s] %s\n", timestamp, msg)
}

func PrintBanner() {
	fmt.Println("------------------------------------------------------------")
	fmt.Println(`
 _____ _____ ____   ____     _____                       _
|  ___|_   _|  _ \ / ___|   | ____|_  ___ __   ___  _ __| |_ ___ _ __
| |_    | | | | | | |       |  _| \ \/ / '_ \ / _ \| '__| __/ _ \ '__|
|  _|   | | | |_| | |___    | |___ >  <| |_) | (_) | |  | ||  __/ |
|_|     |_| |____/ \____|   |_____/_/\_\ .__/ \___/|_|   \__\___|_|
                                       |_|
`)
}
