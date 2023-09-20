package config

import (
	"fmt"

	"github.com/docker/libtrust"
	"github.com/portward/registry-auth/auth"
	"github.com/portward/registry-auth/auth/token/jwt"
	"gopkg.in/yaml.v3"
)

// RefreshTokenIssuerFactory creates a new [auth.RefreshTokenIssuer].
type RefreshTokenIssuerFactory = Factory[auth.RefreshTokenIssuer]

var refreshTokenIssuerFactoryRegistry = &factoryRegistry[auth.RefreshTokenIssuer]{}

// RegisterRefreshTokenIssuerFactory makes a [RefreshTokenIssuerFactory] available by the provided name in configuration.
//
// If RegisterRefreshTokenIssuerFactory is called twice with the same name or if factory is nil, it panics.
func RegisterRefreshTokenIssuerFactory(name string, factory func() RefreshTokenIssuerFactory) {
	err := refreshTokenIssuerFactoryRegistry.RegisterFactory(name, factory)
	if err != nil {
		panic("registering refresh token issuer factory: " + err.Error())
	}
}

func init() {
	RegisterRefreshTokenIssuerFactory("jwt", func() RefreshTokenIssuerFactory { return jwtRefreshTokenIssuer{} })
}

// RefreshTokenIssuer is the configuration for an auth.RefreshTokenIssuer.
type RefreshTokenIssuer struct {
	RefreshTokenIssuerFactory
}

func (c *RefreshTokenIssuer) UnmarshalYAML(value *yaml.Node) error {
	var rawConfig rawConfig

	err := value.Decode(&rawConfig)
	if err != nil {
		return err
	}

	factory, ok := refreshTokenIssuerFactoryRegistry.GetFactory(rawConfig.Type)
	if !ok {
		c.RefreshTokenIssuerFactory = unknownFactoryType[auth.RefreshTokenIssuer]{
			factoryType: "refresh token issuer",
			typ:         rawConfig.Type,
		}

		return nil
	}

	err = decode(rawConfig.Config, &factory)
	if err != nil {
		return err
	}

	c.RefreshTokenIssuerFactory = factory

	return nil
}

type jwtRefreshTokenIssuer struct {
	Issuer         string `mapstructure:"issuer"`
	PrivateKeyFile string `mapstructure:"privateKeyFile"`
}

func (c jwtRefreshTokenIssuer) New() (auth.RefreshTokenIssuer, error) {
	signingKey, err := libtrust.LoadKeyFile(c.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	return jwt.NewRefreshTokenIssuer(c.Issuer, signingKey), nil
}

func (c jwtRefreshTokenIssuer) Validate() error {
	if c.Issuer == "" {
		return fmt.Errorf("jwt: issuer is required")
	}

	if c.PrivateKeyFile == "" {
		return fmt.Errorf("jwt: privateKeyFile is required")
	}

	return nil
}
