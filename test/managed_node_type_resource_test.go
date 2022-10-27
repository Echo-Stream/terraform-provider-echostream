package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccManagedNodeTypeResource(t *testing.T) {
	t.Parallel()
	name := "test"
	readme := "This is a readme"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccManagedNodeTypeResourceConfig(
					"one",
					"226390263822.dkr.ecr.us-east-1.amazonaws.com/hl7-mllp-inbound-node:0.5-dev",
					name,
					nil,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_managed_node_type.test", "description", "one"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "echostream_managed_node_type.test",
				ImportState:       true,
				ImportStateId:     name,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccManagedNodeTypeResourceConfig(
					"two",
					"226390263822.dkr.ecr.us-east-1.amazonaws.com/hl7-mllp-inbound-node:0.5-dev",
					name,
					&readme,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_managed_node_type.test", "description", "two"),
					resource.TestCheckResourceAttr("echostream_managed_node_type.test", "readme", readme),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccManagedNodeTypeResourceConfig(
	description string,
	imageUri string,
	name string,
	readme *string,
) string {
	additonal_params := ""
	if readme != nil {
		additonal_params += fmt.Sprintf(`
  readme = %q`, *readme)
	}
	return fmt.Sprintf(`
resource "echostream_managed_node_type" "test" {
  description = %[1]q
  image_uri = %[2]q
  name = %[3]q
  mount_requirements= [
    {
      description = "my mount1"
      target      = "/foo"
    }
  ]
  port_requirements = [
    {
      container_port = 2575
      description    = "HL7 MLLP"
      protocol       = "tcp"
    }
  ]%[4]s
}
`, description, imageUri, name, additonal_params)
}
