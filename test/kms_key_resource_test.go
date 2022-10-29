package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccKmsKeyResource(t *testing.T) {
	t.Parallel()
	name := "test"
	description := "two"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKmsKeyResourceConfig(
					nil,
					name,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_kms_key.test", "name", name),
				),
			},
			// ImportState testing
			{
				ResourceName:      "echostream_kms_key.test",
				ImportState:       true,
				ImportStateId:     name,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccKmsKeyResourceConfig(
					&description,
					name,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_kms_key.test", "description", description),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccKmsKeyResourceConfig(
	description *string,
	name string,
) string {
	additonal_params := ""
	if description != nil {
		additonal_params += fmt.Sprintf(`
  description = %q`, *description)
	}
	return fmt.Sprintf(`
resource "echostream_kms_key" "test" {
  name = %[1]q%[2]s
}
`, name, additonal_params)
}
