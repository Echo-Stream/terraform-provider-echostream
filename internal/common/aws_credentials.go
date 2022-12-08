package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
	accessKeyId string,
	expiration string,
	secretAccessKey string,
	sessionToken string,
) map[string]attr.Value {
	return map[string]attr.Value{
		"access_key_id":     types.StringValue(accessKeyId),
		"expiration":        types.StringValue(expiration),
		"secret_access_key": types.StringValue(secretAccessKey),
		"session_token":     types.StringValue(sessionToken),
	}
}

func AwsCredentialsSchema() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		"access_key_id": dsschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The AWS Acces Key Id for the session.",
		},
		"expiration": dsschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The date/time that the sesssion expires, in [ISO8601](https://en.wikipedia.org/wiki/ISO_8601) format.",
		},
		"secret_access_key": dsschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The AWS Secret Access Key for the session.",
			Sensitive:           true,
		},
		"session_token": dsschema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "The AWS Session Token for the session.",
		},
	}
}
