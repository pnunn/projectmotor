package github

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var Config = &oauth2.Config{
	ClientID:     "0e590762d1dc627e801f",
	ClientSecret: "fb369c7863ea205ab3a897f3af66986f9f83b30f",
	Scopes:       []string{"read:user", "user:email"},
	Endpoint:     github.Endpoint,
}
