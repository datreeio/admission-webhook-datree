package loggerUtil

import (
	"fmt"
	"os"

	"github.com/datreeio/admission-webhook-datree/pkg/enums"
)

func Log(msg string) {
	fmt.Println(msg)
}

func Logf(format string, a ...any) {
	fmt.Println(fmt.Sprintf(format, a...))
}

func Debug(msg string) {
	if os.Getenv(enums.Debug) == "true" {
		fmt.Println(msg)
	}
}

func Debugf(format string, a ...any) {
	if os.Getenv(enums.Debug) == "true" {
		fmt.Println(fmt.Sprintf(format, a...))
	}
}
