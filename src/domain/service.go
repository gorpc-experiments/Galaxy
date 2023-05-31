package domain

// A Host represents the pod that is running the service
type Host struct {
	Name     string
	LastSeen int64
}

// An Instance represents a running service, it can have multiple hosts (pods) that are running the same service
type Instance struct {
	Address string
	Hosts   []Host
}

// A Service represents a function of a module that can be called remotely like creating a user, saving a specific data, etc
type Service struct {
	Name     string
	Instance Instance
}

// A Module represents a collection of services, it's a major component of the system like Authentication, Notifications, Storage, etc...
// it can have multiple services
type Module struct {
	Name string

	Service []Service
}

// Example of a module

// var Authentication = Module{
// 	Name:        "authentication",
// 	Service: []Service{
// 		{
// 			Name: "login",
// 			Instance: Instance{
// 				Address: "127.0.0.1:50051",
// 				Hosts: []Host{
// 					{
// 						Name:     "pod1",
// 						LastSeen: 0,
// 					},
// 				},
// 			},
// 		},
// 	},
// }