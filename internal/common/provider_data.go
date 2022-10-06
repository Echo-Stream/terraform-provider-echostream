package common

import "github.com/Khan/genqlient/graphql"

type ProviderData struct {
	// graphql client used to make API calls to EchoStream
	Client graphql.Client

	//EchoStream Tenant that this provider is for
	Tenant string
}
