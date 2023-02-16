package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccWebSubHubNodeResource(t *testing.T) {
	t.Parallel()
	name := "hubtest"
	subscriptionSecurity := "httpsandsecret"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccWebSubHubNodeResourceConfig(
					"one",
					name,
					nil,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_web_sub_hub_node.test", "description", "one"),
					resource.TestCheckResourceAttr("echostream_web_sub_hub_node.test", "signature_algorithm", "sha1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "echostream_web_sub_hub_node.test",
				ImportState:       true,
				ImportStateId:     name,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccWebSubHubNodeResourceConfig(
					"two",
					name,
					&subscriptionSecurity,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_web_sub_hub_node.test", "description", "two"),
					resource.TestCheckResourceAttr("echostream_web_sub_hub_node.test", "subscription_security", subscriptionSecurity),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})

}

func testAccWebSubHubNodeResourceConfig(
	description string,
	name string,
	subscriptionSecurity *string,
) string {
	additonal_params := ""
	if subscriptionSecurity != nil {
		additonal_params = fmt.Sprintf(`
  subscription_security = %q`, *subscriptionSecurity)
	}
	return fmt.Sprintf(`
resource "echostream_web_sub_hub_node" "test" {
	description = %[1]q
	name = %[2]q
}
`, description, name, additonal_params)
}
