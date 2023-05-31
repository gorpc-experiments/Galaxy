package service

import (
	"errors"
	"github.com/gorpc-experiments/galaxy/src/domain"
	"github.com/rs/zerolog/log"
	"strings"
)

func (t *Galaxy) LookUp(args *domain.LookUpRequest, quo *domain.LookUpResponse) error {
	log.Info().Str("method", args.ServiceMethod).Msg("Looking up method...")

	method := strings.Split(args.ServiceMethod, ".")

	if len(method) != 2 {
		err := errors.New("Invalid method name")

		log.Err(err).Msg("Unable to look up method")
		return err
	}

	module := method[0]
	service := method[1]

	for _, v := range t.ServiceLibrary {
		if v.Name == module {
			for _, s := range v.Service {
				if s.Name == service {
					quo.Address = s.Instance.Address
					log.Info().Str("method", args.ServiceMethod).Str("at", quo.Address).Msg("Method found")
					return nil
				}
			}
		}
	}
	return nil
}