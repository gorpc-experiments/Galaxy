package main

import (
	"fmt"
	"github.com/AliceDiNunno/KubernetesUtil"
	"github.com/davecgh/go-spew/spew"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
)

type LookUpRequest struct {
	ServiceMethod string
}

type LookUpResponse struct {
	Address string
}

type RegisterRequest struct {
	//Component any
	Address    string
	Components []string
	Service    any
}

type RegisterResponse struct {
	Success bool
}

type AvailableServices []string

type ServiceLibrary map[string]AvailableServices

type Galaxy int

func (t *Galaxy) LookUp(args *LookUpRequest, quo *LookUpResponse) error {
	println("LookUp: ", args.ServiceMethod)
	for k, v := range serviceLibrary {
		for service := range v {
			if v[service] == args.ServiceMethod {
				quo.Address = k
				println("Location: ", quo.Address)
				return nil
			}
		}
	}

	return nil
}

var serviceLibrary ServiceLibrary

func (t *Galaxy) Register(args *RegisterRequest, quo *RegisterResponse) error {
	spew.Dump("Register: ", args.Components, " at ", args.Address)

	if serviceLibrary == nil {
		serviceLibrary = make(ServiceLibrary)
	}

	serviceLibrary[args.Address] = args.Components

	spew.Dump(args.Service)

	return nil
}

func getPort() int {
	port := 0
	if KubernetesUtil.IsRunningInKubernetes() {
		port = KubernetesUtil.GetInternalServicePort()
	}
	if port == 0 {
		env_port := os.Getenv("PORT")
		if env_port == "" {
			log.Fatalln("PORT env variable isn't set")
		}
		envport, err := strconv.Atoi(env_port)
		if err != nil {
			log.Fatalln(err.Error())
		}
		port = envport
	}

	return port
}

func main() {
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		fmt.Println(pair[0], "=", pair[1])
	}

	arith := new(Galaxy)
	err := rpc.Register(arith)
	if err != nil {
		log.Println(err.Error())
		return
	}

	rpc.HandleHTTP()

	port := getPort()

	println("Galaxy is running on port", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Println(err.Error())
	}
}
