package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTenantDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccTenantDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.echostream_tenant.test", "active", "true"),
					resource.TestCheckResourceAttr("data.echostream_tenant.test", "region", "us-east-1"),
				),
			},
		},
	})
}

const testAccTenantDataSourceConfig = `
data "echostream_tenant" "test" {
}
`
