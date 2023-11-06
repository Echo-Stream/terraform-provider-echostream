package app

import (
	"regexp"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"golang.org/x/exp/maps"
)

func appDataSourceAttributes() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"description": dataSourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "A human-readable description of the app.",
		},
		"name": dataSourceSchema.StringAttribute{
			MarkdownDescription: "The name of the app; must be unique in the Tenant.",
			Required:            true,
		},
	}
}

func appResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"description": resourceSchema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "A human-readable description of the app.",
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "The name of the app; must be unique in the Tenant.",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Required:            true,
			Validators:          common.NameValidators,
		},
	}
}

func remoteAppDataSourceAttributes() map[string]dataSourceSchema.Attribute {
	attributes := appDataSourceAttributes()
	maps.Copy(
		attributes,
		map[string]dataSourceSchema.Attribute{
			"audit_records_endpoint": dataSourceSchema.StringAttribute{
				Computed: true,
				MarkdownDescription: "The app-specific endpoint for posting audit records. Details about this endpoint may be found" +
					" [here](https://docs.echo.stream/docs/auditing-messages-from-cross-accountexternalmanaged-apps#auditing-without-use-of-the-echostreamnode-package).",
			},
			"config": dataSourceSchema.StringAttribute{
				Computed:            true,
				CustomType:          common.ConfigType{},
				MarkdownDescription: "The config for the app. All nodes in the app will be allowed to access this. Must be a JSON object.",
				Sensitive:           true,
			},
			"credentials": dataSourceSchema.SingleNestedAttribute{
				Attributes:          common.CognitoCredentialsDataSourceSchema(),
				Computed:            true,
				MarkdownDescription: "The AWS Cognito Credentials that allow the app to access the EchoStream GraphQL API.",
			},
			"table_access": dataSourceSchema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Indicates if this app can gain access to the Tenant's DynamoDB [table](https://docs.echo.stream/docs/table).",
			},
		},
	)
	return attributes
}

func remoteAppResourceAttributes() map[string]resourceSchema.Attribute {
	attributes := appResourceAttributes()
	maps.Copy(
		attributes,
		map[string]resourceSchema.Attribute{
			"audit_records_endpoint": resourceSchema.StringAttribute{
				Computed: true,
				MarkdownDescription: "The app-specific endpoint for posting audit records. Details about this endpoint may be found" +
					" [here](https://docs.echo.stream/docs/auditing-messages-from-cross-accountexternalmanaged-apps#auditing-without-use-of-the-echostreamnode-package).",
			},
			"config": resourceSchema.StringAttribute{
				CustomType:          common.ConfigType{},
				MarkdownDescription: "The config for the app. All nodes in the app will be allowed to access this. Must be a JSON object.",
				Optional:            true,
				Sensitive:           true,
			},
			"credentials": resourceSchema.SingleNestedAttribute{
				Attributes:          common.CognitoCredentialsResourceSchema(),
				Computed:            true,
				MarkdownDescription: "The AWS Cognito Credentials that allow the app to access the EchoStream GraphQL API.",
			},
			"table_access": resourceSchema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Indicates if this app can gain access to the Tenant's DynamoDB [table](https://docs.echo.stream/docs/table).",
				Optional:            true,
			},
		},
	)
	return attributes
}

func managedAppInstanceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"app": resourceSchema.StringAttribute{
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
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "The name of the instance data generated. Changing this is the mechanism for regenerating instance data.",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			Required:            true,
		},
	}
}
