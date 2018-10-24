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
	succes         bool
	log            []string
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
		if ui.succes {
			tm.Print(tm.Color(tm.Bold("Compiles With Warnings:"), tm.YELLOW) + "\n")
		} else {
			tm.Print(tm.Color(tm.Bold("Failed to compile:"), tm.YELLOW) + "\n")
		}
	} else {
		tm.Print(tm.Color(tm.Bold("Compiled successfully."), tm.GREEN) + "\n\n")

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

//ShowBuiltWarning should be shown when something build correctly but has warnings
func (ui *UI) ShowBuiltWarning(warn Warning, page string, line string, data []interface{}) {
	tm.Clear()
	tm.MoveCursor(1, 1)

	tm.Print("" +
		tm.Color(tm.Bold("Compiled with warnings."), tm.YELLOW) + "\n" +
		"\n" +
		tm.Background(tm.Color(tm.Bold(page), tm.BLACK), tm.WHITE) + "\n" +
		"   Line " + line + ":  " + fmt.Sprintf(warn.extensive, data) + "   " + tm.Color(warn.short, tm.YELLOW) + "\n" +
		"")

	tm.Flush()

}

//showError should be shown when something errors out
func (ui *UI) showError(err Error, page string, line string, data []interface{}) {
	tm.Clear()
	tm.MoveCursor(1, 1)

	tm.Print("" +
		tm.Color(tm.Bold("Failed to compile."), tm.RED) + "\n" +
		"\n" +
		tm.Background(tm.Color(tm.Bold(page), tm.BLACK), tm.WHITE) + "\n" +
		"   Line " + line + ":  " + fmt.Sprintf(err.extensive, data) + "   " + tm.Color(err.short, tm.RED) + "\n" +
		"")

	tm.Flush()
}

func getIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
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

func (ui *UI) Info(err string) {
	entry := tm.Color(tm.Bold("Info:."), tm.BLUE) + err + "\n"
	ui.log = append(ui.log, entry)
	ui.LogFile.Write([]byte(entry + "\n"))
}

func (ui *UI) Error(err string) {
	entry := tm.Color(tm.Bold("Error:."), tm.YELLOW) + err + "\n"
	ui.log = append(ui.log, entry)
	ui.LogFile.Write([]byte(entry + "\n"))

}

func (ui *UI) Fatal(err string) {
	entry := tm.Color(tm.Bold("Failled to compile:"), tm.RED) + err + "\n"
	ui.log = append(ui.log, entry)
	ui.LogFile.Write([]byte(entry + "\n"))

	ui.succes = false
}
