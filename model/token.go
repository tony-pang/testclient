package model

type Token struct {
	UserID       string    `yaml:"userId"`
	IDToken      string    `yaml:"idToken"`
	SessionToken string    `yaml:"sessionToken"`
	ExpiresIn    int       `yaml:"expiresIn"`
	User         TokenUser `yaml:"user"`
}

type TokenUser struct {
	ID       string `yaml:"id"`
	IDDomain string `yaml:"idDomain"`
}
