package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSystemMessageTypeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccSystemMessageTypeDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.echostream_system_message_type.test", "name", "Change Emitter"),
				),
			},
		},
	})
}

const testAccSystemMessageTypeDataSourceConfig = `
data "echostream_system_message_type" "test" {
}
`
