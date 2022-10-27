package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTenantResource(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTenantResourceConfig("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_tenant.test", "description", "one"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "echostream_tenant.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccTenantResourceConfig("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_tenant.test", "description", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTenantResourceConfig(description string) string {
	return fmt.Sprintf(`
resource "echostream_tenant" "test" {
  description = %[1]q
}
`, description)
}
