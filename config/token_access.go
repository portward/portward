package config

import (
	"fmt"
	"time"

	"github.com/docker/libtrust"
	"github.com/portward/registry-auth/auth"
	"github.com/portward/registry-auth/auth/token/jwt"
	"gopkg.in/yaml.v3"
)

// AccessTokenIssuerFactory creates a new [auth.AccessTokenIssuer].
type AccessTokenIssuerFactory = Factory[auth.AccessTokenIssuer]

var accessTokenIssuerFactoryRegistry = &factoryRegistry[auth.AccessTokenIssuer]{}

// RegisterAccessTokenIssuerFactory makes a [AccessTokenIssuerFactory] available by the provided name in configuration.
//
// If RegisterAccessTokenIssuerFactory is called twice with the same name or if factory is nil, it panics.
func RegisterAccessTokenIssuerFactory(name string, factory func() AccessTokenIssuerFactory) {
	err := accessTokenIssuerFactoryRegistry.RegisterFactory(name, factory)
	if err != nil {
		panic("registering access token issuer factory: " + err.Error())
	}
}

func init() {
	RegisterAccessTokenIssuerFactory("jwt", func() AccessTokenIssuerFactory { return jwtAccessTokenIssuer{} })
}

// AccessTokenIssuer is the configuration for an auth.AccessTokenIssuer.
type AccessTokenIssuer struct {
	AccessTokenIssuerFactory
}

func (c *AccessTokenIssuer) UnmarshalYAML(value *yaml.Node) error {
	var rawConfig rawConfig

	err := value.Decode(&rawConfig)
	if err != nil {
		return err
	}

	factory, ok := accessTokenIssuerFactoryRegistry.GetFactory(rawConfig.Type)
	if !ok {
		c.AccessTokenIssuerFactory = unknownFactoryType[auth.AccessTokenIssuer]{
			factoryType: "access token issuer",
			typ:         rawConfig.Type,
		}

		return nil
	}

	err = decode(rawConfig.Config, &factory)
	if err != nil {
		return err
	}

	c.AccessTokenIssuerFactory = factory

	return nil
}

type jwtAccessTokenIssuer struct {
	Issuer         string        `mapstructure:"issuer"`
	PrivateKeyFile string        `mapstructure:"privateKeyFile"`
	Expiration     time.Duration `mapstructure:"expiration"`
}

func (c jwtAccessTokenIssuer) New() (auth.AccessTokenIssuer, error) {
	signingKey, err := libtrust.LoadKeyFile(c.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	return jwt.NewAccessTokenIssuer(c.Issuer, signingKey, c.Expiration), nil
}

func (c jwtAccessTokenIssuer) Validate() error {
	if c.Issuer == "" {
		return fmt.Errorf("jwt: issuer is required")
	}

	if c.PrivateKeyFile == "" {
		return fmt.Errorf("jwt: privateKeyFile is required")
	}

	if c.Expiration == 0 {
		return fmt.Errorf("jwt: expiration is required")
	}

	return nil
}
