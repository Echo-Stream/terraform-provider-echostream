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

type appModel struct {
	Description types.String `tfsdk:"description"`
	Name        types.String `tfsdk:"name"`
}

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

type crossAccountAppModel struct {
	Account              types.String  `tfsdk:"account"`
	AppsyncEndpoint      types.String  `tfsdk:"appsync_endpoint"`
	AuditRecordsEndpoint types.String  `tfsdk:"audit_records_endpoint"`
	Config               common.Config `tfsdk:"config"`
	Credentials          types.Object  `tfsdk:"credentials"`
	Description          types.String  `tfsdk:"description"`
	IamPolicy            types.String  `tfsdk:"iam_policy"`
	Name                 types.String  `tfsdk:"name"`
	TableAccess          types.Bool    `tfsdk:"table_access"`
}

type crossTenantReceivingAppModel struct {
	Description   types.String `tfsdk:"description"`
	Name          types.String `tfsdk:"name"`
	SendingApp    types.String `tfsdk:"sending_app"`
	SendingTenant types.String `tfsdk:"sending_tenant"`
}

type crossTenantSendingAppModel struct {
	Description     types.String `tfsdk:"description"`
	Name            types.String `tfsdk:"name"`
	ReceivingApp    types.String `tfsdk:"receiving_app"`
	ReceivingTenant types.String `tfsdk:"receiving_tenant"`
}

type externalAppModel struct {
	AppsyncEndpoint      types.String  `tfsdk:"appsync_endpoint"`
	AuditRecordsEndpoint types.String  `tfsdk:"audit_records_endpoint"`
	Config               common.Config `tfsdk:"config"`
	Credentials          types.Object  `tfsdk:"credentials"`
	Description          types.String  `tfsdk:"description"`
	Name                 types.String  `tfsdk:"name"`
	TableAccess          types.Bool    `tfsdk:"table_access"`
}

type managedAppModel struct {
	AuditRecordsEndpoint types.String  `tfsdk:"audit_records_endpoint"`
	Config               common.Config `tfsdk:"config"`
	Credentials          types.Object  `tfsdk:"credentials"`
	Description          types.String  `tfsdk:"description"`
	Name                 types.String  `tfsdk:"name"`
	TableAccess          types.Bool    `tfsdk:"table_access"`
}

func crossTenantReceivingSchema() map[string]tfsdk.Attribute {
	schema := appSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"sending_app": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Optional:            true,
				Type:                types.StringType,
			},
			"sending_tenant": {
				Description:         "",
				MarkdownDescription: "",
				Required:            true,
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Type:                types.StringType,
			},
		},
	)
	description := schema["description"]
	description.Computed = true
	return schema
}

func crossTenantSendingSchema() map[string]tfsdk.Attribute {
	schema := appSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"receiving_app": {
				Description:         "",
				MarkdownDescription: "",
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Required:            true,
				Type:                types.StringType,
			},
			"receiving_tenant": {
				Description:         "",
				MarkdownDescription: "",
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Required:            true,
				Type:                types.StringType,
			},
		},
	)
	description := schema["description"]
	description.Computed = true
	return schema
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

func crossAccountAppSchema() map[string]tfsdk.Attribute {
	schema := remoteAppSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"account": {
				Description:         "",
				MarkdownDescription: "",
				PlanModifiers:       tfsdk.AttributePlanModifiers{resource.RequiresReplace()},
				Required:            true,
				Type:                types.StringType,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.LengthBetween(12, 12),
					stringvalidator.RegexMatches(
						regexp.MustCompile("^[0-9]+$"),
						"value must contain only numbers",
					),
				},
			},
			"appsync_endpoint": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
			"iam_policy": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
		},
	)
	return schema
}

func externalAppSchema() map[string]tfsdk.Attribute {
	schema := remoteAppSchema()
	maps.Copy(
		schema,
		map[string]tfsdk.Attribute{
			"appsync_endpoint": {
				Computed:            true,
				Description:         "",
				MarkdownDescription: "",
				Type:                types.StringType,
			},
		},
	)
	return schema
}

func managedAppSchema() map[string]tfsdk.Attribute {
	schema := remoteAppSchema()
	name := schema["name"]
	name.Validators = []tfsdk.AttributeValidator{
		stringvalidator.LengthBetween(3, 80),
		stringvalidator.RegexMatches(
			regexp.MustCompile(`^[A-Za-z0-9\-\_]*$`),
			"value must contain only lowercase/uppercase alphanumeric characters, \"-\", \"_\"",
		),
	}
	return schema
}
