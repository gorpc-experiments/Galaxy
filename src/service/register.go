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

	//found := false

	for serviceIndex := range t.ServiceLibrary {
		//Already registered services
		service := t.ServiceLibrary[serviceIndex]

		moduleName := service.Name
		moduleToRegister := t.getModule(args)

		//If the module is already registered we will update the address
		if moduleName == moduleToRegister {
			moduleServices := service.Service
			componentsToRegister := args.Components

			for componentIndex := range moduleServices {
				component := moduleServices[componentIndex]

				//If the component is already registered we will update the address
				for _, componentToRegister := range componentsToRegister {
					if component.Name == componentToRegister {
						t.ServiceLibrary[serviceIndex].Service[componentIndex].Instance.Address = args.Address

						//Add host or update last seen

						foundHost := false
						for host := range t.ServiceLibrary[serviceIndex].Service[componentIndex].Instance.Hosts {
							if t.ServiceLibrary[serviceIndex].Service[componentIndex].Instance.Hosts[host].Name == args.Host {
								t.ServiceLibrary[serviceIndex].Service[componentIndex].Instance.Hosts[host].LastSeen = time.Now().Unix()
								foundHost = true
							}
						}

						if !foundHost {
							t.ServiceLibrary[serviceIndex].Service[componentIndex].Instance.Hosts = append(t.ServiceLibrary[serviceIndex].Service[componentIndex].Instance.Hosts, domain.Host{
								Name:     args.Host,
								LastSeen: time.Now().Unix(),
							})
						}

						log.Info().Strs("components", args.Components).Str("address", args.Address).Msg("Replaced existing service")
					}
				}
			}
		}
	}

	if !t.moduleExists(t.getModule(args)) {
		//Create new module
		t.ServiceLibrary = append(t.ServiceLibrary, domain.Module{
			Name:    t.getModule(args),
			Service: []domain.Service{},
		})
	}

	for _, component := range args.Components {
		currentModule := t.getModule(args)

		found := false

		for serviceIndex := range t.ServiceLibrary {
			service := t.ServiceLibrary[serviceIndex]

			if service.Name == currentModule {
				for componentIndex := range service.Service {
					componentToCheck := service.Service[componentIndex]

					if componentToCheck.Name == component {
						found = true
					}
				}
			}
		}

		//some components like health are used internally and should not be registered as a service
		if !found {
			if !t.isModuleBlacklisted(component) {
				t.ServiceLibrary = append(t.ServiceLibrary, domain.Module{
					Name: currentModule,
					Service: []domain.Service{
						{
							Name: component,
							Instance: domain.Instance{
								Address: args.Address,

								Hosts: []domain.Host{
									{
										Name:     args.Host,
										LastSeen: time.Now().Unix(),
									},
								},
							},
						},
					},
				})
			}
		}
	}

	spew.Dump(t.ServiceLibrary)

	return nil
}