package wf

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			err := recover()
			if err != nil {
				log.Printf("%s\n\n", trace(fmt.Sprintf("%s", err)))
				// log.Printf("%s\n\n", trace(err.(string)))
				c.JSON(http.StatusInternalServerError, H{"error": "Internal Server Error"})
			}
		}()

		c.Next()
	}
}

func trace(msg string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(msg + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}
