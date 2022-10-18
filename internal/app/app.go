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
			Description:         "",
			MarkdownDescription: "",
			Type:                types.StringType,
		},
		"name": {
			Description:         "",
			MarkdownDescription: "",
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
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
			"config": {
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Sensitive:           true,
				Type:                common.ConfigType{},
			},
			"credentials": {
				Attributes:          tfsdk.SingleNestedAttributes(common.CognitoCredentialsSchema()),
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
			},
			"table_access": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
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
			Description:         "",
			MarkdownDescription: "",
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
			Description:         "",
			MarkdownDescription: "",
			PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
			Required:            true,
			Type:                types.StringType,
		},
	}
}
