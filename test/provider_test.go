package provider_test

import (
	"os"
	"testing"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"echostream": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	suffixes := []string{"APPSYNC_ENDPOINT", "CLIENT_ID", "PASSWORD", "TENANT", "USERNAME", "USER_POOL_ID"}
	for _, keySuffix := range suffixes {
		require.NotEmpty(t, os.Getenv("ECHOSTREAM_"+keySuffix), "ECHOSTREAM_"+keySuffix+" must be set")
	}
}
