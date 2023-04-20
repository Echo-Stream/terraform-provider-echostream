package test

import (
	"encoding/json"
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
				Config: testAccTenantResourceConfig("one", map[string]any{"foo": "bar"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_tenant.test", "description", "one"),
					resource.TestCheckResourceAttr("echostream_tenant.test", "config", "{\"foo\":\"bar\"}"),
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
				Config: testAccTenantResourceConfig("two", map[string]any{"foo": []int{1, 8, 10, 32, 2, 9, 15, 80, 1001, 0}}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_tenant.test", "description", "two"),
					resource.TestCheckResourceAttr("echostream_tenant.test", "config", "{\"foo\":[1,8,10,32,2,9,15,80,1001,0]}"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTenantResourceConfig(description string, config map[string]any) string {
	b, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(`
resource "echostream_tenant" "test" {
  description = %[1]q
  config = %[2]q
}
`, description, string(b))
}
