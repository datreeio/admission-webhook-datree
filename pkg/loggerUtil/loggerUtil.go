package loggerUtil

import (
	"fmt"
	"os"

	"github.com/datreeio/admission-webhook-datree/pkg/enums"
)

func Log(msg string) {
	fmt.Println(msg)
}

func Debug(msg string) {
	if os.Getenv(enums.Debug) == "true" {
		fmt.Println(msg)
	}
}
