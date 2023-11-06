package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dataSourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CognitoCredentialsModel struct {
	ClientId   types.String `tfsdk:"client_id"`
	Password   types.String `tfsdk:"password"`
	UserPoolId types.String `tfsdk:"user_pool_id"`
	Username   types.String `tfsdk:"username"`
}

func CognitoCredentialsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"client_id":    types.StringType,
		"password":     types.StringType,
		"user_pool_id": types.StringType,
		"username":     types.StringType,
	}
}

func CognitoCredentialsAttrValues(
	clientId string,
	password string,
	userPoolId string,
	username string,
) map[string]attr.Value {
	return map[string]attr.Value{
		"client_id":    types.StringValue(clientId),
		"password":     types.StringValue(password),
		"user_pool_id": types.StringValue(userPoolId),
		"username":     types.StringValue(username),
	}
}

func CognitoCredentialsDataSourceSchema() map[string]dataSourceSchema.Attribute {
	return map[string]dataSourceSchema.Attribute{
		"client_id": resourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The AWS Cognito Client ID used to connect to EchoStream.",
		},
		"password": dataSourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The password to use when connecting to EchoStream.",
			Sensitive:           true,
		},
		"user_pool_id": dataSourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The AWS Cognito User Pool ID used to connect to EchoStream.",
		},
		"username": dataSourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The username to use when connecting to EchoStream.",
		},
	}
}

func CognitoCredentialsResourceSchema() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"client_id": resourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The AWS Cognito Client ID used to connect to EchoStream.",
		},
		"password": resourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The password to use when connecting to EchoStream.",
			Sensitive:           true,
		},
		"user_pool_id": resourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The AWS Cognito User Pool ID used to connect to EchoStream.",
		},
		"username": resourceSchema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The username to use when connecting to EchoStream.",
		},
	}
}
