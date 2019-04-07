// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package ui

import (
	"fmt"
	"io"
	"net"

	tm "github.com/buger/goterm"
)

//UI is the way to display stuff on the console.
type UI struct {
	LogFile        io.Writer
	HostingEnabled bool
	Port           string
	failed         bool
	log            []string
	infolog        []string
	PrettyLog      bool
	DebugEnabled   bool
}

var (
	debugPrefix = tm.Color(tm.Bold("debug "), tm.WHITE)
	infoPrefix  = tm.Color(tm.Bold("info "), tm.BLUE)
	errorPrefix = tm.Color(tm.Bold("error "), tm.RED)
	fatalPrefix = tm.Color(tm.Bold("Failed to build. \n\n"), tm.RED)
)

//ShowResult should be shown when something builds successfully
func (ui *UI) ShowResult() {
	tm.Clear()
	tm.MoveCursor(1, 1)

	if len(ui.log) != 0 {
		if ui.failed {
			tm.Print(tm.Color(tm.Bold("Failed to build."), tm.RED) + "\n")
		} else {
			tm.Print(tm.Color(tm.Bold("Built with warnings."), tm.YELLOW) + "\n")
		}
	} else {
		if ui.failed {
			tm.Print(tm.Color(tm.Bold("Failed to build."), tm.RED) + "\n")
		} else {
			tm.Print(tm.Color(tm.Bold("Built successfully."), tm.GREEN) + "\n\n")
			if !ui.HostingEnabled {
				tm.Print("" +
					"The " + tm.Color(tm.Bold("build"), tm.BLUE) + " folder is ready to be deployed.\n" +
					"You should serve it with a static hosting server.\n" +
					"")
			} else {
				tm.Print("" +
					"You can now view your project in the browser.\n" +
					"   Local:            http://localhost:" + ui.Port + "/\n" +
					"   On Your Network:  http://" + getIP() + ":" + ui.Port + "/\n" +
					"\n" +
					"Note that the development build is not optimized. \n" +
					"To create a production build, use " + tm.Color(tm.Bold("antibuild build"), tm.BLUE) + ".\n" +
					"\n" +
					"To rebuild press " + tm.Color(tm.Bold("r"), tm.CYAN) + " and to hard reload press " + tm.Color(tm.Bold("R"), tm.CYAN) + ". Press " + tm.Color(tm.Bold("ESC"), tm.CYAN) + " to quit.\n" +
					"")
			}
		}
	}

	if len(ui.log) != 0 {
		if ui.failed {
			tm.Print("\n" +
				tm.Color(tm.Bold("The following errors have occured:"), tm.RED) + "\n\n")
		} else {
			tm.Print("\n" +
				tm.Color(tm.Bold("The following warnings have been reported:"), tm.BLUE) + "\n\n")
		}
		for _, e := range ui.log { //e for entry
			tm.Print(e + "\n")
		}
		ui.log = make([]string, 0)
	}

	if len(ui.infolog) != 0 {
		tm.Print("\n" +
			tm.Color(tm.Bold("Log:"), tm.BLUE) + "\n")

		for _, e := range ui.infolog { //e for entry
			tm.Print(e + "\n")
		}

		ui.infolog = make([]string, 0)
	}

	tm.Flush()
	ui.failed = false
}

func (ui *UI) showlog() {
	tm.Clear()
	tm.MoveCursor(1, 1)

	tm.Print(tm.Color(tm.Bold("Log:"), tm.BLUE) + "\n")

	for _, e := range ui.infolog { //e for entry
		tm.Print(e + "\n")
	}
	tm.Flush()
}

func getIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

//Debug logs to the log file only
func (ui *UI) Debug(err string) {
	if ui.LogFile != nil || !ui.DebugEnabled {
		if ui.PrettyLog {
			entry := debugPrefix + err
			ui.LogFile.Write([]byte(entry + "\n"))
			return
		}
		ui.LogFile.Write([]byte(err + "\n"))
	}
}

//Debugf logs to the log file only
func (ui *UI) Debugf(format string, a ...interface{}) {
	if !ui.DebugEnabled {
		return
	}

	ui.Debug(fmt.Sprintf(format, a...))
}

//Info logs helpfull information/warnings
func (ui *UI) Info(err string) {
	entry := infoPrefix + err
	ui.infolog = append(ui.infolog, entry)
	if ui.LogFile != nil {
		if ui.PrettyLog {
			ui.LogFile.Write([]byte(entry + "\n"))
			return
		}
		ui.LogFile.Write([]byte(err + "\n"))
	}
	ui.showlog()
}

//Infof logs helpfull information/warnings
func (ui *UI) Infof(format string, a ...interface{}) {
	ui.Info(fmt.Sprintf(format, a...))
}

//Error logs errors, these can later be followed up on with a fatal or have potential consequences for the outcome
func (ui *UI) Error(err string) {
	entry := errorPrefix + err
	ui.log = append(ui.log, entry)
	if ui.LogFile != nil {
		if ui.PrettyLog {
			ui.LogFile.Write([]byte(entry))
			return
		}
		ui.LogFile.Write([]byte(err + "\n"))
	}
}

//Errorf logs errors, these can later be followed up on with a fatal or have potential consequences for the outcome
func (ui *UI) Errorf(format string, a ...interface{}) {
	ui.Error(fmt.Sprintf(format, a...))
}

//Fatal should be called when in an unrecoverable state. EG: config file not found, template function not called etc.
func (ui *UI) Fatal(err string) {
	entry := fatalPrefix + err
	ui.log = append(ui.log, entry)
	ui.failed = true

	if ui.LogFile != nil {
		if ui.PrettyLog {
			ui.LogFile.Write([]byte(entry))
			return
		}
		ui.LogFile.Write([]byte(err + "\n"))
	}
}

//Fatalf should be called when in an unrecoverable state. EG: config file not found, template function not called etc.
func (ui *UI) Fatalf(format string, a ...interface{}) {
	ui.Fatal(fmt.Sprintf(format, a...))
}

//SetLogfile sets the output writer for the logger
func (ui *UI) SetLogfile(file io.Writer) {
	if file != nil {
		ui.DebugEnabled = true
		ui.LogFile = file
	} else {
		ui.DebugEnabled = false
		ui.LogFile = nil
	}
}

//SetPrettyPrint this sets the setting for pretty printing,
func (ui *UI) SetPrettyPrint(enabled bool) {
	ui.PrettyLog = enabled
}

//ShouldEnableDebug enables/disables debug logging (performance)
func (ui *UI) ShouldEnableDebug(enabled bool) {
	ui.DebugEnabled = enabled
}
