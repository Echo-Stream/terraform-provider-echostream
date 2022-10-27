package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccMessageTypeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccMessageTypeDataSourceConfig("echo.json"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.echostream_message_type.test", "name", "echo.json"),
					resource.TestCheckResourceAttr("data.echostream_message_type.test", "requirements.0", "simplejson"),
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
