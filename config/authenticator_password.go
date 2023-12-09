package config

import (
	"fmt"
	"maps"

	"github.com/portward/registry-auth/auth"
	"github.com/portward/registry-auth/auth/authn"
	"github.com/portward/unkey"
	"gopkg.in/yaml.v3"

	"github.com/portward/portward/pkg/slices"
)

// PasswordAuthenticatorFactory creates a new [auth.PasswordAuthenticator].
type PasswordAuthenticatorFactory = Factory[auth.PasswordAuthenticator]

var passwordAuthenticatorFactoryRegistry = &factoryRegistry[auth.PasswordAuthenticator]{}

// RegisterPasswordAuthenticatorFactory makes a [PasswordAuthenticatorFactory] available by the provided name in configuration.
//
// If RegisterPasswordAuthenticatorFactory is called twice with the same name or if factory is nil, it panics.
func RegisterPasswordAuthenticatorFactory(name string, factory func() PasswordAuthenticatorFactory) {
	err := passwordAuthenticatorFactoryRegistry.RegisterFactory(name, factory)
	if err != nil {
		panic("registering password authenticator factory: " + err.Error())
	}
}

func init() {
	RegisterPasswordAuthenticatorFactory("user", func() PasswordAuthenticatorFactory { return userAuthenticator{} })
	RegisterPasswordAuthenticatorFactory("unkey", func() PasswordAuthenticatorFactory { return unkey.Config{} })
}

// PasswordAuthenticator is the configuration for an [auth.PasswordAuthenticator].
type PasswordAuthenticator struct {
	PasswordAuthenticatorFactory
}

func (c *PasswordAuthenticator) UnmarshalYAML(value *yaml.Node) error {
	var rawConfig rawConfig

	err := value.Decode(&rawConfig)
	if err != nil {
		return err
	}

	factory, ok := passwordAuthenticatorFactoryRegistry.GetFactory(rawConfig.Type)
	if !ok {
		c.PasswordAuthenticatorFactory = unknownFactoryType[auth.PasswordAuthenticator]{
			factoryType: "password authenticator",
			typ:         rawConfig.Type,
		}

		return nil
	}

	err = decode(rawConfig.Config, &factory)
	if err != nil {
		return err
	}

	c.PasswordAuthenticatorFactory = factory

	return nil
}

type userAuthenticator struct {
	Entries []user `mapstructure:"entries"`
}

type user struct {
	Enabled      bool           `mapstructure:"enabled"`
	Username     string         `mapstructure:"username"`
	PasswordHash string         `mapstructure:"passwordHash"`
	Attrs        map[string]any `mapstructure:"attributes"`
}

func (c userAuthenticator) New() (auth.PasswordAuthenticator, error) {
	entries := slices.Map(c.Entries, func(v user) authn.User {
		return authn.User{
			Enabled:      v.Enabled,
			Username:     v.Username,
			PasswordHash: v.PasswordHash,
			Attrs:        maps.Clone(v.Attrs),
		}
	})

	return authn.NewUserAuthenticator(entries), nil
}

func (c userAuthenticator) Validate() error {
	for i, entry := range c.Entries {
		if entry.Username == "" {
			return fmt.Errorf("user authenticator: entry[%d]: username is required", i)
		}

		if entry.PasswordHash == "" {
			return fmt.Errorf("user authenticator: entry[%d]: password hash is required", i)
		}
	}

	return nil
}
