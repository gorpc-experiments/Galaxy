package service

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorpc-experiments/galaxy/src/domain"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

func (t *Galaxy) hasService(serviceList []string, service string) bool {
	for _, s := range serviceList {
		if s == service {
			return true
		}
	}

	return false
}

func (t *Galaxy) checkApiVersion(args *domain.RegisterRequest) error {
	if args.Version != 1 {
		mismatchedVersionError := errors.New("API Version mismatch")
		log.Err(mismatchedVersionError).Strs("components", args.Components).Str("address", args.Address).Int("current", 1).Int("remote", args.Version).Msg("Unable to register components")
		return mismatchedVersionError
	}

	return nil
}

func (t *Galaxy) getModule(args *domain.RegisterRequest) string {
	if len(args.Components) > 1 {
		firstComponent := args.Components[0]
		split := strings.Split(firstComponent, ".")
		if len(split) == 2 {
			return split[0]
		}

	}

	return ""
}

func (t *Galaxy) getService(name string) string {
	split := strings.Split(name, ".")
	if len(split) == 2 {
		return split[1]
	}
	return ""
}

func (t *Galaxy) moduleExists(module string) bool {
	for _, m := range t.ServiceLibrary {
		if m.Name == module {
			return true
		}
	}

	return false
}

func (t *Galaxy) checkArguments(args *domain.RegisterRequest) error {
	if len(args.Components) == 0 {
		err := errors.New("no components to register")

		log.Err(err).Strs("components", args.Components).Str("address", args.Address).Msg("Unable to register components")
		return err
	}

	module := ""

	for _, component := range args.Components {
		if strings.Contains(component, ".") {
			currentModule := strings.Split(component, ".")[0]

			if module == "" {
				module = currentModule
			} else if module != currentModule {
				err := errors.New("multiple module exports are not supported")

				log.Err(err).Strs("components", args.Components).Str("address", args.Address).Msg("Unable to register components")
				return err
			}
		}
	}
	return nil
}

func (t *Galaxy) checkForMandatoryExport(args *domain.RegisterRequest) error {
	module := t.getModule(args)

	if !t.hasService(args.Components, fmt.Sprintf("%s.%s", module, "Health")) {
		err := errors.New("health function not exported")

		log.Err(err).Strs("components", args.Components).Str("address", args.Address).Msg("Unable to register components. Did you inherit from CoreHealth ?")
		return err
	}

	return nil
}

func (t *Galaxy) isModuleBlacklisted(name string) bool {
	blacklistedValues := []string{"Health"}
	return t.hasService(blacklistedValues, t.getService(name))
}

func (t *Galaxy) registerModule(module string) {
	if !t.moduleExists(module) {
		t.ServiceLibrary = append(t.ServiceLibrary, domain.Module{
			Name:    module,
			Service: []domain.Service{},
		})
	}
}

func (t *Galaxy) registerService(module string, component string, host string, address string) {
	split := strings.Split(component, ".")
	if len(split) != 2 {
		return
	}

	for moduleIndex, m := range t.ServiceLibrary {
		if m.Name == module {
			serviceFound := false
			for serviceIndex, s := range m.Service {
				if s.Name == split[1] {
					t.ServiceLibrary[moduleIndex].Service[serviceIndex].Instance.Address = address
					hostFound := false
					for hostIndex, h := range s.Instance.Hosts {
						if h.Name == host {
							t.ServiceLibrary[moduleIndex].Service[serviceIndex].Instance.Hosts[hostIndex].LastSeen = time.Now().Unix()
							hostFound = true
						}
					}
					if !hostFound {
						t.ServiceLibrary[moduleIndex].Service[serviceIndex].Instance.Hosts = append(t.ServiceLibrary[moduleIndex].Service[serviceIndex].Instance.Hosts, domain.Host{
							Name:     host,
							LastSeen: time.Now().Unix(),
						})
					}
					serviceFound = true
					break
				}
			}
			if !serviceFound {
				t.ServiceLibrary[moduleIndex].Service = append(t.ServiceLibrary[moduleIndex].Service, domain.Service{
					Name: split[1],
					Instance: domain.Instance{
						Address: address,
						Hosts: []domain.Host{
							{
								Name:     host,
								LastSeen: time.Now().Unix(),
							},
						},
					},
				})
			}
		}
	}
}

func (t *Galaxy) register(args *domain.RegisterRequest) {
	host := args.Host
	address := args.Address

	if len(args.Components) > 0 {
		split := strings.Split(args.Components[0], ".")
		module := ""
		if len(split) == 2 {
			module = split[0]
		}

		if !t.moduleExists(module) {
			t.registerModule(module)
		}

		for _, component := range args.Components {
			if !t.isModuleBlacklisted(component) {
				t.registerService(module, component, host, address)
			}
		}
	}
}

func (t *Galaxy) Register(args *domain.RegisterRequest, quo *domain.RegisterResponse) error {
	log.Info().Strs("components", args.Components).Str("address", args.Address).Msg("Registering new components")

	if err := t.checkApiVersion(args); err != nil {
		log.Info().Strs("components", args.Components).Str("address", args.Address).Msg(err.Error())
		return err
	}

	if err := t.checkArguments(args); err != nil {
		log.Info().Strs("components", args.Components).Str("address", args.Address).Msg(err.Error())

		return err
	}

	if err := t.checkForMandatoryExport(args); err != nil {
		return err
	}

	t.register(args)

	spew.Dump(t.ServiceLibrary)

	return nil
}