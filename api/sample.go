package api

import (
	"fmt"
	"strings"
)

func SampleRPC(bearer string) {
	fmt.Println("Starting RPC sample")

	fmt.Println("Initial files :")
	PrintRPCFile(bearer)
	CleanRPC(bearer)

	callbacks := map[string]func(*RPCRequest) RPCReturn{
		"sayhello": func(r *RPCRequest) RPCReturn {
			return RPCReturn{"Hello ! " + strings.Join(r.Parameters, "//"), "you", 666}
		},
		"selfkill": func(r *RPCRequest) RPCReturn {
			// here, remove, kill, destroy and vanish
			// return value will be never used
			return RPCReturn{"", "", 0}
		},
	}

	ratId := "aabbccdd"
	fmt.Println("Write the rpcrequest")
	requestid := WriteRPCRequest(bearer, ratId, &RPCRequest{"sayhello", []string{"1", "2", "3"}})

	fmt.Println("files :")
	PrintRPCFile(bearer)

	fmt.Println("Process the rpcrequest")
	ProcessRPCRequest(bearer, ratId, callbacks)

	fmt.Println("files :")
	PrintRPCFile(bearer)

	fmt.Println("Read the response")
	responses := ReadRPCResponse(bearer)

	fmt.Println("files :")
	PrintRPCFile(bearer)

	for r, resps := range responses {
		for _, ret := range resps {
			fmt.Println("Rat ", r, " return ", ret)
			if requestid == ret.RequestId {
				fmt.Println("Found the request id")
			}
		}
	}

	fmt.Println("end of RPC sample")

}

func SampleRWOnboard(bearer string) {
	fmt.Println("Starting onboard sample")

	data := OnboardData{"hostname", "user"}
	ratId := "aabbccdd"
	fmt.Println("write : ", ratId, data)
	WriteOnboard(bearer, ratId, &data)

	fmt.Println("read onboard")
	myRats := ReadOnboard(bearer)
	fmt.Println("read : ", myRats)
	for i, d := range myRats {
		fmt.Println(i, d)
	}

	fmt.Println("end of onboard sample")

}
