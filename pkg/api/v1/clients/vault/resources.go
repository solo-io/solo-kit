package vault

// Resources represents the raw data models returned by the Vault API

type Certificate map[string]string

type Secret struct {
	Certificate
}

type SecretList []Secret
