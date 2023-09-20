package config

import (
	cerbosauthorizer "github.com/portward/cerbos-authorizer"
	"github.com/portward/registry-auth/auth"
	"github.com/portward/registry-auth/auth/authz"
	"gopkg.in/yaml.v3"
)

// AuthorizerFactory creates a new [auth.Authorizer].
type AuthorizerFactory = Factory[auth.Authorizer]

var authorizerFactoryRegistry = &factoryRegistry[auth.Authorizer]{}

// RegisterAuthorizerFactory makes a [AuthorizerFactory] available by the provided name in configuration.
//
// If RegisterAuthorizerFactory is called twice with the same name or if factory is nil, it panics.
func RegisterAuthorizerFactory(name string, factory func() AuthorizerFactory) {
	err := authorizerFactoryRegistry.RegisterFactory(name, factory)
	if err != nil {
		panic("registering authorizer factory: " + err.Error())
	}
}

func init() {
	RegisterAuthorizerFactory("default", func() AuthorizerFactory { return defaultAuthorizer{} })
	RegisterAuthorizerFactory("cerbos", func() AuthorizerFactory { return cerbosauthorizer.Config{} })
}

// Authorizer is the configuration for an auth.Authorizer.
type Authorizer struct {
	AuthorizerFactory
}

func (c *Authorizer) UnmarshalYAML(value *yaml.Node) error {
	var rawConfig rawConfig

	err := value.Decode(&rawConfig)
	if err != nil {
		return err
	}

	factory, ok := authorizerFactoryRegistry.GetFactory(rawConfig.Type)
	if !ok {
		c.AuthorizerFactory = unknownFactoryType[auth.Authorizer]{
			factoryType: "authorizer",
			typ:         rawConfig.Type,
		}

		return nil
	}

	err = decode(rawConfig.Config, &factory)
	if err != nil {
		return err
	}

	c.AuthorizerFactory = factory

	return nil
}

type defaultAuthorizer struct {
	AllowAnonymous bool `mapstructure:"allowAnonymous"`
}

func (c defaultAuthorizer) New() (auth.Authorizer, error) {
	return authz.NewDefaultAuthorizer(authz.NewDefaultRepositoryAuthorizer(c.AllowAnonymous), c.AllowAnonymous), nil
}

func (c defaultAuthorizer) Validate() error {
	return nil
}
