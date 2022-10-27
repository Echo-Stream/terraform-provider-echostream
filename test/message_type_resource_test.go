package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccMessageTypeResource(t *testing.T) {
	t.Parallel()
	name := "test"
	readme := "This is a readme"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMessageTypeResourceConfig(
					"def auditor(*, message, **kwargs):\n    return {}\n",
					"def bitmapper(*, context, message, source, **kwargs):\n\n    bitmap = 0x0\n\n    # TODO - Perform conditional bitmapping of the text message here.\n    # The returned bitmap will be compared against the routes in the\n    # route table for route matching. This will be done by\n    # (message_bitmap \u0026 route) == route. If no routes match, the\n    # message will be filtered.\n\n    return bitmap\n",
					"one",
					name,
					"def processor(*, context, message, source, **kwargs):\n\n    # TODO - Perform any transformations to the message here.\n    # This can include transforming the message to something that\n    # is no longer text. Remember, you MUST return a string or None.\n    # If None is returned, the message will be filtered.\n\n    return message\n",
					"Processing Platform as a Service (pPaaS) is a suite of cloud services enabling development, execution and governance of processing flows connecting any combination of on premises and cloud-based processes, services, applications and data within individual or across multiple organizations.",
					nil,
					[]string{},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_message_type.test", "description", "one"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "echostream_message_type.test",
				ImportState:       true,
				ImportStateId:     name,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccMessageTypeResourceConfig(
					"def auditor(*, message, **kwargs):\n    return {}\n",
					"def bitmapper(*, context, message, source, **kwargs):\n\n    bitmap = 0x0\n\n    # TODO - Perform conditional bitmapping of the text message here.\n    # The returned bitmap will be compared against the routes in the\n    # route table for route matching. This will be done by\n    # (message_bitmap \u0026 route) == route. If no routes match, the\n    # message will be filtered.\n\n    return bitmap\n",
					"two",
					name,
					"def processor(*, context, message, source, **kwargs):\n\n    # TODO - Perform any transformations to the message here.\n    # This can include transforming the message to something that\n    # is no longer text. Remember, you MUST return a string or None.\n    # If None is returned, the message will be filtered.\n\n    return message\n",
					"Processing Platform as a Service (pPaaS) is a suite of cloud services enabling development, execution and governance of processing flows connecting any combination of on premises and cloud-based processes, services, applications and data within individual or across multiple organizations.",
					&readme,
					[]string{"requests", "simplejson"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("echostream_message_type.test", "description", "two"),
					resource.TestCheckResourceAttr("echostream_message_type.test", "readme", readme),
					resource.TestCheckResourceAttr("echostream_message_type.test", "requirements.0", "requests"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccMessageTypeResourceConfig(
	auditor string,
	bitmapperTemplate string,
	description string,
	name string,
	processorTemplate string,
	sampleMessage string,
	readme *string,
	requirements []string,
) string {
	additonal_params := ""
	if readme != nil {
		additonal_params += fmt.Sprintf(`
  readme = %q`, *readme)
	}
	if len(requirements) > 0 {
		additonal_params += fmt.Sprintf(`
  requirements = ["%s"]`, strings.Join(requirements, `", "`))
	}
	return fmt.Sprintf(`
resource "echostream_message_type" "test" {
  auditor = %[1]q
  bitmapper_template = %[2]q
  description = %[3]q
  name = %[4]q
  processor_template = %[5]q
  sample_message = %[6]q%[7]s
}
`, auditor, bitmapperTemplate, description, name, processorTemplate, sampleMessage, additonal_params)
}
