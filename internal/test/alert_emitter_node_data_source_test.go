package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAlertEmitterNodeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccAlertEmitterNodeDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.echostream_alert_emitter_node.test", "name", "Alert Emitter"),
				),
			},
		},
	})
}

const testAccAlertEmitterNodeDataSourceConfig = `
data "echostream_alert_emitter_node" "test" {
}
`
