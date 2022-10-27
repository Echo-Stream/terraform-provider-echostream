package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AwsCredentialsModel struct {
	AccessKeyId     types.String `tfsdk:"access_key_id"`
	Expiration      types.String `tfsdk:"expiration"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	SessionToken    types.String `tfsdk:"session_token"`
}

func AwsCredentialsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"access_key_id":     types.StringType,
		"expiration":        types.StringType,
		"secret_access_key": types.StringType,
		"session_token":     types.StringType,
	}
}

func AwsCredentialsAttrValues(
	accessKeyIdd string,
	expiration string,
	secretAccessKey string,
	sessionToken string,
) map[string]attr.Value {
	return map[string]attr.Value{
		"access_key_id":     types.String{Value: accessKeyIdd},
		"expiration":        types.String{Value: expiration},
		"secret_access_key": types.String{Value: secretAccessKey},
		"session_token":     types.String{Value: sessionToken},
	}
}

func AwsCredentialsSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"access_key_id": {
			Computed:            true,
			MarkdownDescription: "The AWS Acces Key Id for the session.",
			Type:                types.StringType,
		},
		"expiration": {
			Computed:            true,
			MarkdownDescription: "The date/time that the sesssion expires, in [ISO8601](https://en.wikipedia.org/wiki/ISO_8601) format.",
			Type:                types.StringType,
		},
		"secret_access_key": {
			Computed:            true,
			MarkdownDescription: "The AWS Secret Access Key for the session.",
			Sensitive:           true,
			Type:                types.StringType,
		},
		"session_token": {
			Computed:            true,
			MarkdownDescription: "The AWS Session Token for the session.",
			Type:                types.StringType,
		},
	}
}
