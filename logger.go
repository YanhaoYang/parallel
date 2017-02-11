package parallel

import (
	"log"
	"os"
)

type simpleLogger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Fatalln(v ...interface{})
}

func initLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)
}
