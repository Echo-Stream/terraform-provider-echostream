package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccMessageTypeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMessageTypeResourceConfig("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_message_type.test", "description", "one"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "echostream_message_type.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"config", "description"},
			},
			// Update and Read testing
			{
				Config: testAccMessageTypeResourceConfig("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_message_type.test", "description", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccMessageTypeResourceConfig(description string) string {
	return fmt.Sprintf(`
resource "echostream_message_type" "test" {
  description = %[1]q
}
`, description)
}
