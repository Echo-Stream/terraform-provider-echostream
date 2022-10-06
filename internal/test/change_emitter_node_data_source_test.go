package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccChangeEmitterNodeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccChangeEmitterNodeDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.echostream_change_emitter_node.test", "name", "Change Emitter"),
				),
			},
		},
	})
}

const testAccChangeEmitterNodeDataSourceConfig = `
data "echostream_change_emitter_node" "test" {
}
`
