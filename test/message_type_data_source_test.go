package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccMessageTypeDataSource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccMessageTypeDataSourceConfig("echo.hl7"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.echostream_message_type.test", "name", "echo.hl7"),
					resource.TestCheckResourceAttr("data.echostream_message_type.test", "requirements.0", "hl7>=0.4.2"),
				),
			},
		},
	})
}

func testAccMessageTypeDataSourceConfig(name string) string {
	return fmt.Sprintf(`
data "echostream_message_type" "test" {
	name = %[1]q
}
`, name)
}
