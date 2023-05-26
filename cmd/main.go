package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorpc-experiments/ServiceCore"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	log.Info().Str("method", args.ServiceMethod).Msg("Looking up method...")
	for k, v := range serviceLibrary {
		for service := range v {
			if v[service] == args.ServiceMethod {
				quo.Address = k
				log.Info().Str("method", args.ServiceMethod).Str("at", quo.Address).Msg("Method found")
				return nil
			}
		}
	}

	return nil
}

var serviceLibrary ServiceLibrary

func (t *Galaxy) Register(args *RegisterRequest, quo *RegisterResponse) error {
	log.Info().Strs("components", args.Components).Str("address", args.Address).Msg("Registering new components")

	if serviceLibrary == nil {
		serviceLibrary = make(ServiceLibrary)
	}

	serviceLibrary[args.Address] = args.Components

	spew.Dump(args.Service)

	return nil
}

func setupLogging() {
	if os.Getenv("ENV") == "debug" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Short caller (file:line)
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		if strings.HasPrefix(file, "/app/") {
			file = strings.TrimPrefix(file, "/app/")
		}
		return file + ":" + strconv.Itoa(line)
	}

	log.Logger = log.With().Caller().Logger()
}

func main() {
	setupLogging()

	log.Debug().Strs("env", os.Environ()).Msg("Starting")

	arith := new(Galaxy)
	err := rpc.Register(arith)
	if err != nil {
		log.Err(err)
		return
	}

	rpc.HandleHTTP()

	port := ServiceCore.GetRPCPort()

	log.Info().Int("port", port).Msg("Galaxy is running")
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Err(err)
	}
}
