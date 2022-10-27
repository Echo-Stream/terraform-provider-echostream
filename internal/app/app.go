package app

import (
	"regexp"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/maps"
)

func appSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"description": {
			Optional:            true,
			MarkdownDescription: "A human-readable description of the app.",
			Type:                types.StringType,
		},
		"name": {
			MarkdownDescription: "The name of the app; must be unique in the Tenant.",
			PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			Required:            true,
			Type:                types.StringType,
			Validators:          common.NameValidators,
		},
	}
}

func remoteAppSchema() map[string]tfsdk.Attribute {
	schema := appSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"audit_records_endpoint": {
				Computed: true,
				MarkdownDescription: "The app-specific endpoint for posting audit records. Details about this endpoint may be found" +
					" [here](https://docs.echo.stream/docs/auditing-messages-from-cross-accountexternalmanaged-apps#auditing-without-use-of-the-echostreamnode-package).",
				Type: types.StringType,
			},
			"config": {
				MarkdownDescription: "The config for the app. All nodes in the app will be allowed to access this. Must be a JSON object.",
				Optional:            true,
				Sensitive:           true,
				Type:                common.ConfigType{},
			},
			"credentials": {
				Attributes:          tfsdk.SingleNestedAttributes(common.CognitoCredentialsSchema()),
				Computed:            true,
				MarkdownDescription: "The AWS Cognito Credentials that allow the app to access the EchoStream GraphQL API.",
			},
			"table_access": {
				Computed:            true,
				MarkdownDescription: "Indicates if this app can gain access to the Tenant's DynamoDB [table](https://docs.echo.stream/docs/table).",
				Optional:            true,
				Type:                types.BoolType,
			},
		},
	)
	return schema
}

func managedAppInstanceSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"app": {
			MarkdownDescription: "The name of the app.",
			PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			Required:            true,
			Type:                types.StringType,
			Validators: []tfsdk.AttributeValidator{
				stringvalidator.LengthBetween(3, 80),
				stringvalidator.RegexMatches(
					regexp.MustCompile(`^[A-Za-z0-9\-\_]*$`),
					"value must contain only lowercase/uppercase alphanumeric characters, \"-\", \"_\"",
				),
			},
		},
		"name": {
			MarkdownDescription: "The name of the instance data generated. Changing this is the mechanism for regenerating instance data.",
			PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			Required:            true,
			Type:                types.StringType,
		},
	}
}
