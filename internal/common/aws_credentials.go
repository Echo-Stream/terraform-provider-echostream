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
	access_key_id string,
	expiration string,
	secret_access_key string,
	session_token string,
) map[string]attr.Value {
	return map[string]attr.Value{
		"access_key_id":     types.String{Value: access_key_id},
		"expiration":        types.String{Value: expiration},
		"secret_access_key": types.String{Value: secret_access_key},
		"session_token":     types.String{Value: session_token},
	}
}

func AwsCredentialsSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"access_key_id": {
			Computed: true,
			Type:     types.StringType,
		},
		"expiration": {
			Computed: true,
			Type:     types.StringType,
		},
		"secret_access_key": {
			Computed:  true,
			Sensitive: true,
			Type:      types.StringType,
		},
		"session_token": {
			Computed: true,
			Type:     types.StringType,
		},
	}
}
