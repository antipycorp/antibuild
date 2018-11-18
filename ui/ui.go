package ui

import (
	"fmt"
	"io"
	"net"

	tm "github.com/buger/goterm"
)

//Error s are a set of predefined errors to be displayed in the ui
type Error struct {
	extensive string
	short     string
}

//Warning s are set of predefined warnings to be displayed in the ui
type Warning struct {
	extensive string
	short     string
}

//UI is the way to display stuff on the console.
type UI struct {
	LogFile        io.Writer
	HostingEnabled bool
	Port           string
	failed         bool
	log            []string
	PrettyLog      bool
}

//ShowCompiling should be shown when you start compiling
func (ui *UI) ShowCompiling() {
	//tm.Clear()
	//tm.MoveCursor(1, 1)

	tm.Print("" +
		"Compiling...\n" +
		"\n")

	tm.Flush()
}

//ShowResult should be shown when something builds successfully
func (ui *UI) ShowResult() {
	tm.Clear()
	tm.MoveCursor(1, 1)
	if len(ui.log) != 0 {
		if ui.failed {
			tm.Print(tm.Color(tm.Bold("Failed to compile:"), tm.RED) + "\n")
		} else {
			tm.Print(tm.Color(tm.Bold("Compiled With Warnings:"), tm.YELLOW) + "\n")
		}
	} else {
		if ui.failed {
			tm.Print(tm.Color(tm.Bold("Failed to compile:"), tm.RED) + "\n")
		} else {
			tm.Print(tm.Color(tm.Bold("Compiled successfully."), tm.GREEN) + "\n\n")
		}
	}
	if !ui.HostingEnabled {
		tm.Print("" +
			"The " + tm.Color(tm.Bold("build"), tm.BLUE) + " folder is ready to be deployed.\n" +
			"You should serve it with a static server.\n" +
			"")
	} else {
		tm.Print("" +
			"You can now view your project in the browser.\n" +
			"   Local:            http://localhost:" + ui.Port + "/\n" +
			"   On Your Network:  http://" + getIP() + ":" + ui.Port + "/\n" +
			"\n" +
			"Note that the development build is not optimized. \n" +
			"To create a production build, use " + tm.Color(tm.Bold("antibuild build"), tm.BLUE) + ".\n" +
			"")
	}
	if len(ui.log) != 0 {
		tm.Print("\n" +
			tm.Color(tm.Bold("the following errors have occured:"), tm.YELLOW) + "\n")
		for _, e := range ui.log { //e for entry
			tm.Print(e + "\n")
		}
	}

	tm.Flush()
}

func getIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "failled to get IP address"
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
	if ui.LogFile != nil {
		if ui.PrettyLog {
			entry := tm.Color(tm.Bold("Debug:."), tm.WHITE) + err
			ui.LogFile.Write([]byte(entry + "\n"))
			return
		}
		ui.LogFile.Write([]byte(err + "\n"))
	}
}

//Debugf logs to the log file only
func (ui *UI) Debugf(format string, a ...interface{}) {
	ui.Debug(fmt.Sprintf(format, a...))
}

//Info logs helpfull information/warnings
func (ui *UI) Info(err string) {
	entry := tm.Color(tm.Bold("Info:."), tm.BLUE) + err
	ui.log = append(ui.log, entry)
	if ui.LogFile != nil {
		if ui.PrettyLog {
			ui.LogFile.Write([]byte(entry))
			return
		}
		ui.LogFile.Write([]byte(err + "\n"))
	}
}

//Infof logs helpfull information/warnings
func (ui *UI) Infof(format string, a ...interface{}) {
	ui.Info(fmt.Sprintf(format, a...))
}

//Error logs errors, these can later be followed up on with a fatal or have potential consequences for the outcome
func (ui *UI) Error(err string) {
	entry := tm.Color(tm.Bold("Error:."), tm.YELLOW) + err
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
	entry := tm.Color(tm.Bold("Failled to compile:."), tm.RED) + err
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

func (ui *UI) Setlogfile(file io.Writer) {
	ui.LogFile = file
}
func (ui *UI) Setprettyprint(enabled bool) {
	ui.PrettyLog = enabled
}
