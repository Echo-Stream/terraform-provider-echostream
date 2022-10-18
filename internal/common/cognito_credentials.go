package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
		"client_id":    types.String{Value: clientId},
		"password":     types.String{Value: password},
		"user_pool_id": types.String{Value: userPoolId},
		"username":     types.String{Value: username},
	}
}

func CognitoCredentialsSchema() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"client_id": {
			Computed: true,
			Type:     types.StringType,
		},
		"password": {
			Computed:  true,
			Sensitive: true,
			Type:      types.StringType,
		},
		"user_pool_id": {
			Computed: true,
			Type:     types.StringType,
		},
		"username": {
			Computed: true,
			Type:     types.StringType,
		},
	}
}
