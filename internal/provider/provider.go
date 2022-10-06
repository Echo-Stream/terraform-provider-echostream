package provider

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/function"
	kmskey "github.com/Echo-Stream/terraform-provider-echostream/internal/kms_key"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/message_type"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/node"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/tenant"
	"github.com/Khan/genqlient/graphql"
	cognitosrp "github.com/alexrudd/cognito-srp/v4"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	cognitoIdp "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitoIdp_types "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// Ensure EchoStreamDoer satifies doer interface
var _ graphql.Doer = &echoStreamApiDoer{}

// Ensure EchoStreamProvider satisfies various provider interfaces.
var _ provider.ProviderWithMetadata = &echoStreamProvider{}

type echoStreamApiDoer struct {
	sync.Mutex
	accessToken  *string
	cidp         *cognitoIdp.Client
	clientId     string
	expiration   time.Time
	refreshToken *string
}

func (d *echoStreamApiDoer) getToken(ctx context.Context) (*string, error) {
	if time.Now().After(d.expiration) {
		d.Lock()
		defer d.Unlock()
		// Access token has expired, refresh
		if time.Now().After(d.expiration) {
			resp, err := d.cidp.InitiateAuth(
				ctx,
				&cognitoIdp.InitiateAuthInput{
					AuthFlow: cognitoIdp_types.AuthFlowTypeRefreshTokenAuth,
					AuthParameters: map[string]string{
						string(cognitoIdp_types.AuthFlowTypeRefreshToken): *d.refreshToken,
					},
					ClientId: aws.String(d.clientId),
				},
			)
			if err != nil {
				return nil, err
			}
			token, err := jwt.ParseInsecure([]byte(*resp.AuthenticationResult.AccessToken))
			if err != nil {
				return nil, err
			}
			d.accessToken = resp.AuthenticationResult.AccessToken
			d.refreshToken = resp.AuthenticationResult.RefreshToken
			d.expiration = token.Expiration()
		}
	}
	return d.accessToken, nil
}

func newEchoStreamDoer(ctx context.Context, data *EchoStreamProviderModel) (*echoStreamApiDoer, error) {
	d := echoStreamApiDoer{
		clientId: data.ClientId.Value,
	}
	csrp, err := cognitosrp.NewCognitoSRP(
		data.Username.Value,
		data.Password.Value,
		data.UserPoolId.Value,
		data.ClientId.Value,
		nil,
	)
	if err != nil {
		return nil, err
	}
	// configure cognito identity provider
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(strings.Split(data.UserPoolId.Value, "_")[0]),
	)
	if err != nil {
		return nil, err
	}
	d.cidp = cognitoIdp.NewFromConfig(cfg)
	// initiate auth
	resp, err := d.cidp.InitiateAuth(ctx, &cognitoIdp.InitiateAuthInput{
		AuthFlow:       cognitoIdp_types.AuthFlowTypeUserSrpAuth,
		ClientId:       aws.String(d.clientId),
		AuthParameters: csrp.GetAuthParams(),
	})
	if err != nil {
		return nil, err
	}
	// respond to password verifier challenge
	if resp.ChallengeName == cognitoIdp_types.ChallengeNameTypePasswordVerifier {
		challengeResponses, _ := csrp.PasswordVerifierChallenge(resp.ChallengeParameters, time.Now())

		resp, err := d.cidp.RespondToAuthChallenge(ctx, &cognitoIdp.RespondToAuthChallengeInput{
			ChallengeName:      cognitoIdp_types.ChallengeNameTypePasswordVerifier,
			ChallengeResponses: challengeResponses,
			ClientId:           aws.String(csrp.GetClientId()),
		})
		if err != nil {
			return nil, err
		}

		token, err := jwt.ParseInsecure([]byte(*resp.AuthenticationResult.AccessToken))
		if err != nil {
			return nil, err
		}
		d.accessToken = resp.AuthenticationResult.AccessToken
		d.expiration = token.Expiration()
		d.refreshToken = resp.AuthenticationResult.RefreshToken
		return &d, nil
	}
	return nil, errors.New("Invalid challenge: " + string(resp.ChallengeName))
}

func (d *echoStreamApiDoer) Do(req *http.Request) (*http.Response, error) {
	if token, err := d.getToken(req.Context()); err != nil {
		return nil, err
	} else {
		req.Header.Set("Authorization", *token)
	}
	return http.DefaultClient.Do(req)
}

// echoStreamProvider defines the provider implementation.
type echoStreamProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *echoStreamProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource { return &node.AlertEmitterNodeDataSource{} },
		func() datasource.DataSource { return &function.ApiAuthenticatorFunctionDataSource{} },
		func() datasource.DataSource { return &node.AppChangeRouterNodeDataSource{} },
		func() datasource.DataSource { return &node.AuditEmitterNodeDataSource{} },
		func() datasource.DataSource { return &function.BitmapperFunctionDataSource{} },
		func() datasource.DataSource { return &node.ChangeEmitterNodeDataSource{} },
		func() datasource.DataSource { return &message_type.MessageTypeDataSource{} },
		func() datasource.DataSource { return &function.ProcessorFunctionDataSource{} },
		func() datasource.DataSource { return &tenant.TenantDataSource{} },
	}
}

func (p *echoStreamProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data EchoStreamProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	doer, err := newEchoStreamDoer(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Rrror creating api connection", err.Error())
		return
	}

	// Example client configuration for data sources and resources
	pd := common.ProviderData{
		Client: graphql.NewClient(data.AppsyncEndpoint.Value, doer),
		Tenant: data.Tenant.Value,
	}
	resp.DataSourceData = &pd
	resp.ResourceData = &pd
}

// EchoStreamProviderModel describes the provider data model.
type EchoStreamProviderModel struct {
	AppsyncEndpoint types.String `tfsdk:"appsync_endpoint"`
	ClientId        types.String `tfsdk:"client_id"`
	Password        types.String `tfsdk:"password"`
	Tenant          types.String `tfsdk:"tenant"`
	Username        types.String `tfsdk:"username"`
	UserPoolId      types.String `tfsdk:"user_pool_id"`
}

func (p *echoStreamProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"appsync_endpoint": {
				Description:         "The EchoStream AppSync Endpoint to connect to",
				MarkdownDescription: "The EchoStream AppSync Endpoint to connect to.",
				Required:            true,
				Type:                types.StringType,
			},
			"client_id": {
				Description:         "",
				MarkdownDescription: "",
				Required:            true,
				Type:                types.StringType,
			},
			"password": {
				Description:         "",
				MarkdownDescription: "",
				Required:            true,
				Sensitive:           true,
				Type:                types.StringType,
			},
			"tenant": {
				Description:         "",
				MarkdownDescription: "",
				Required:            true,
				Type:                types.StringType,
			},
			"username": {
				Description:         "",
				MarkdownDescription: "",
				Required:            true,
				Type:                types.StringType,
			},
			"user_pool_id": {
				Description:         "",
				MarkdownDescription: "",
				Required:            true,
				Type:                types.StringType,
			},
		},
		Description:         "",
		MarkdownDescription: "All attributes other than `tenant` must be copied from from the `ApiUser` that you configured in your `Tenant`. The `ApiUser` must have `admin` access to be used in Terraform.",
	}, nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &echoStreamProvider{
			version: version,
		}
	}
}

func (p *echoStreamProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "echostream"
	resp.Version = p.version
}

func (p *echoStreamProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &function.ApiAuthenticatorFunctionResource{} },
		func() resource.Resource { return &kmskey.KmsKeyResource{} },
		func() resource.Resource { return &message_type.MessageTypeResource{} },
		func() resource.Resource { return &function.BitmapperFunctionResource{} },
		func() resource.Resource { return &function.ProcessorFunctionResource{} },
		func() resource.Resource { return &tenant.TenantResource{} },
	}
}
