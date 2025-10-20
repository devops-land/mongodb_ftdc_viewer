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
_____ _____ ____   ____   ____
|  ___|_   _|  _ \ / ___| |  _ \ _ __ ___   ___ ___  ___ ___  ___  _ __
| |_    | | | | | | |     | |_) | '__/ _ \ / __/ _ \/ __/ __|/ _ \| '__|
|  _|   | | | |_| | |___  |  __/| | | (_) | (_|  __/\__ \__ \ (_) | |
|_|     |_| |____/ \____| |_|   |_|  \___/ \___\___||___/___/\___/|_|
`)
}
