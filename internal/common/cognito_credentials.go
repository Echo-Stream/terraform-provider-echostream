package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

func CognitoCredentialsSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"client_id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The AWS Cognito Client ID used to connect to EchoStream.",
		},
		"password": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The password to use when connecting to EchoStream.",
			Sensitive:           true,
		},
		"user_pool_id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The AWS Cognito User Pool ID used to connect to EchoStream.",
		},
		"username": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The username to use when connecting to EchoStream.",
		},
	}
}
