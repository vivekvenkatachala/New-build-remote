package servicelogs

import (
	"fmt"
	"os"

	log "github.com/alecthomas/log4go"
)

type MethodLogs struct {
	Email, Module, MethodName, Description, DetailedDescription string;
	
}


var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func Init() {
	_, err := os.Stat("remote_build_logs")
	if os.IsNotExist(err) {
		err = os.Mkdir("remote_build_logs", 0755)
		if err != nil {
			fmt.Println(err)
		}
	}

	// log.LoadConfiguration("../../log-config.xml")

	// file, err := os.OpenFile("nife-logs/logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	// if err != nil {
	// 	log.Println(err)
	// }

	// InfoLogger = log.Info("INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	// InfoLogger = InfoLogger
	// WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	// ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

