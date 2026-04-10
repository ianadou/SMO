// Package ports defines the interfaces (ports) that the domain layer
// requires from the outside world to function.
//
// Following the hexagonal architecture pattern, the domain owns these
// interfaces and infrastructure adapters implement them. This inversion
// of dependency means the domain never imports infrastructure code.
//
// Ports defined in this package include repositories (for persistence)
// and will later include other infrastructure dependencies such as
// notification senders, ID generators, and clocks.
package ports
