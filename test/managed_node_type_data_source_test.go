package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccManagedNodeTypeDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccManagedNodeTypeDataSourceConfig("echo.hl7-mllp-inbound-node"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.echostream_managed_node_type.test", "name", "echo.hl7-mllp-inbound-node"),
					resource.TestCheckResourceAttr("data.echostream_managed_node_type.test", "send_message_type", "echo.hl7"),
				),
			},
		},
	})
}

func testAccManagedNodeTypeDataSourceConfig(name string) string {
	return fmt.Sprintf(`
data "echostream_managed_node_type" "test" {
	name = %[1]q
}
`, name)
}
