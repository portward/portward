package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestComplete(t *testing.T) {
	file, err := os.Open("testdata/complete.yaml")
	require.NoError(t, err)
	defer file.Close()

	var actual Config

	err = yaml.NewDecoder(file).Decode(&actual)
	if err != nil {
		t.Fatal(err)
	}

	err = actual.Validate()
	if err != nil {
		t.Fatal(err)
	}

	expected := Config{
		PasswordAuthenticator: PasswordAuthenticator{
			PasswordAuthenticatorFactory: userAuthenticator{
				Entries: []user{
					{
						Enabled:      true,
						Username:     "user",
						PasswordHash: "$2a$12$vox7h99HV.gzbZGeBj69jeJVgkkP2nHTndG9USjp..00.WtIqvSpa",
						Attrs: map[string]any{
							"group": "admin",
							"roles": []any{"user", "admin"},
						},
					},
				},
			},
		},
		AccessTokenIssuer: AccessTokenIssuer{
			AccessTokenIssuerFactory: jwtAccessTokenIssuer{
				Issuer:         "localhost:8080",
				PrivateKeyFile: "private_key.pem",
				Expiration:     15 * time.Minute,
			},
		},
		RefreshTokenIssuer: RefreshTokenIssuer{
			RefreshTokenIssuerFactory: jwtRefreshTokenIssuer{
				Issuer:         "localhost:8080",
				PrivateKeyFile: "private_key.pem",
			},
		},
		Authorizer: Authorizer{
			AuthorizerFactory: defaultAuthorizer{
				AllowAnonymous: true,
			},
		},
	}

	assert.Equal(t, expected, actual)
}
