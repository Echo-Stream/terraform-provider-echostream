package app

import (
	"regexp"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"golang.org/x/exp/maps"
)

func appAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"description": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "A human-readable description of the app.",
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "The name of the app; must be unique in the Tenant.",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Required:            true,
			Validators:          common.NameValidators,
		},
	}
}

func remoteAppAttributes() map[string]schema.Attribute {
	attributes := appAttributes()
	maps.Copy(
		attributes,
		map[string]schema.Attribute{
			"audit_records_endpoint": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "The app-specific endpoint for posting audit records. Details about this endpoint may be found" +
					" [here](https://docs.echo.stream/docs/auditing-messages-from-cross-accountexternalmanaged-apps#auditing-without-use-of-the-echostreamnode-package).",
			},
			"config": schema.StringAttribute{
				CustomType:          common.ConfigType{},
				MarkdownDescription: "The config for the app. All nodes in the app will be allowed to access this. Must be a JSON object.",
				Optional:            true,
				Sensitive:           true,
			},
			"credentials": schema.SingleNestedAttribute{
				Attributes:          common.CognitoCredentialsSchema(),
				Computed:            true,
				MarkdownDescription: "The AWS Cognito Credentials that allow the app to access the EchoStream GraphQL API.",
			},
			"table_access": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Indicates if this app can gain access to the Tenant's DynamoDB [table](https://docs.echo.stream/docs/table).",
				Optional:            true,
			},
		},
	)
	return attributes
}

func managedAppInstanceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"app": schema.StringAttribute{
			MarkdownDescription: "The name of the app.",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(3, 80),
				stringvalidator.RegexMatches(
					regexp.MustCompile(`^[A-Za-z0-9\-\_]*$`),
					"value must contain only lowercase/uppercase alphanumeric characters, \"-\", \"_\"",
				),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "The name of the instance data generated. Changing this is the mechanism for regenerating instance data.",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Required:            true,
		},
	}
}
