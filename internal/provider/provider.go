package provider

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Echo-Stream/terraform-provider-echostream/internal/app"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/common"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/edge"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/function"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/kmskey"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/managed_node_type"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/message_type"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/node"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/tenant"
	"github.com/Echo-Stream/terraform-provider-echostream/internal/user"
	"github.com/Khan/genqlient/graphql"
	cognitosrp "github.com/alexrudd/cognito-srp/v4"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	cognitoIdp "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitoIdp_types "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	// Ensure EchoStreamDoer satifies doer interface
	_ graphql.Doer = &echoStreamApiDoer{}

	// Ensure EchoStreamProvider satisfies various provider interfaces.
	_ provider.ProviderWithMetadata = &echoStreamProvider{}
	_ provider.ProviderWithSchema   = &echoStreamProvider{}
)

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
		clientId: data.ClientId.ValueString(),
	}
	csrp, err := cognitosrp.NewCognitoSRP(
		data.Username.ValueString(),
		data.Password.ValueString(),
		data.UserPoolId.ValueString(),
		data.ClientId.ValueString(),
		nil,
	)
	if err != nil {
		return nil, err
	}
	// configure cognito identity provider
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(strings.Split(data.UserPoolId.ValueString(), "_")[0]),
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
		func() datasource.DataSource { return &function.ApiAuthenticatorFunctionDataSource{} },
		func() datasource.DataSource { return &function.BitmapperFunctionDataSource{} },
		func() datasource.DataSource { return &function.ProcessorFunctionDataSource{} },
		func() datasource.DataSource { return &managed_node_type.ManagedNodeTypeDataSource{} },
		func() datasource.DataSource { return &message_type.MessageTypeDataSource{} },
		func() datasource.DataSource { return &node.AlertEmitterNodeDataSource{} },
		func() datasource.DataSource { return &node.AppChangeReceiverNodeDataSource{} },
		func() datasource.DataSource { return &node.AppChangeRouterNodeDataSource{} },
		func() datasource.DataSource { return &node.AuditEmitterNodeDataSource{} },
		func() datasource.DataSource { return &node.ChangeEmitterNodeDataSource{} },
		func() datasource.DataSource { return &node.DeadLetterEmitterNodeDataSource{} },
		func() datasource.DataSource { return &node.LogEmitterNodeDataSource{} },
		func() datasource.DataSource { return &tenant.TenantDataSource{} },
	}
}

func (p *echoStreamProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data EchoStreamProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.AppsyncEndpoint.IsNull() {
		appsyncEndpoint := os.Getenv("ECHOSTREAM_APPSYNC_ENDPOINT")
		if appsyncEndpoint != "" {
			data.AppsyncEndpoint = types.StringValue(appsyncEndpoint)
		} else {
			resp.Diagnostics.AddError(
				"Missing AppSync Endpoint Configuration",
				"While configuring the provider, the AppSync Endpoint was not found in "+
					"the ECHOSTREAM_APPSYNC_ENDPOINT environment variable or provider "+
					"configuration block appsync_endpoint attribute.",
			)
		}
	}
	if data.ClientId.IsNull() {
		clientId := os.Getenv("ECHOSTREAM_CLIENT_ID")
		if clientId != "" {
			data.ClientId = types.StringValue(clientId)
		} else {
			resp.Diagnostics.AddError(
				"Missing Client ID Configuration",
				"While configuring the provider, the AppSync Endpoint was not found in "+
					"the ECHOSTREAM_CLIENT_ID environment variable or provider "+
					"configuration block client_id attribute.",
			)
		}
	}
	if data.Password.IsNull() {
		password := os.Getenv("ECHOSTREAM_PASSWORD")
		if password != "" {
			data.Password = types.StringValue(password)
		} else {
			resp.Diagnostics.AddError(
				"Missing Password Configuration",
				"While configuring the provider, the AppSync Endpoint was not found in "+
					"the ECHOSTREAM_PASSWORD environment variable or provider "+
					"configuration block password attribute.",
			)
		}
	}
	if data.Tenant.IsNull() {
		tenant := os.Getenv("ECHOSTREAM_TENANT")
		if tenant != "" {
			data.Tenant = types.StringValue(tenant)
		} else {
			resp.Diagnostics.AddError(
				"Missing Tenant Configuration",
				"While configuring the provider, the AppSync Endpoint was not found in "+
					"the ECHOSTREAM_TENANT environment variable or provider "+
					"configuration block tenant attribute.",
			)
		}
	}
	if data.Username.IsNull() {
		username := os.Getenv("ECHOSTREAM_USERNAME")
		if username != "" {
			data.Username = types.StringValue(username)
		} else {
			resp.Diagnostics.AddError(
				"Missing Username Configuration",
				"While configuring the provider, the AppSync Endpoint was not found in "+
					"the ECHOSTREAM_USERNAME environment variable or provider "+
					"configuration block username attribute.",
			)
		}
	}
	if data.UserPoolId.IsNull() {
		userPoolId := os.Getenv("ECHOSTREAM_USER_POOL_ID")
		if userPoolId != "" {
			data.UserPoolId = types.StringValue(userPoolId)
		} else {
			resp.Diagnostics.AddError(
				"Missing User Pool ID Configuration",
				"While configuring the provider, the AppSync Endpoint was not found in "+
					"the ECHOSTREAM_USER_POOL_ID environment variable or provider "+
					"configuration block user_pool_id attribute.",
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	doer, err := newEchoStreamDoer(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("Error creating api connection", err.Error())
		return
	}

	// Example client configuration for data sources and resources
	pd := common.ProviderData{
		Client: graphql.NewClient(data.AppsyncEndpoint.ValueString(), doer),
		Tenant: data.Tenant.ValueString(),
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

func (p *echoStreamProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"appsync_endpoint": schema.StringAttribute{
				MarkdownDescription: "The ApiUser's AppSync Endpoint.",
				Optional:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The ApiUser's AWS Cognito Client Id.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The ApiUser's password.",
				Optional:            true,
				Sensitive:           true,
			},
			"tenant": schema.StringAttribute{
				MarkdownDescription: "The EchoStream Tenant to manage.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The ApiUser's username.",
				Optional:            true,
			},
			"user_pool_id": schema.StringAttribute{
				MarkdownDescription: "The ApiUser's AWS Cognito User Pool Id.",
				Optional:            true,
			},
		},
	}
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
		func() resource.Resource { return &app.CrossAccountAppResource{} },
		func() resource.Resource { return &app.CrossTenantReceivingAppResource{} },
		func() resource.Resource { return &app.CrossTenantSendingAppResource{} },
		func() resource.Resource { return &app.ExternalAppResource{} },
		func() resource.Resource { return &app.ManagedAppResource{} },
		func() resource.Resource { return &app.ManagedAppInstanceIsoResource{} },
		func() resource.Resource { return &app.ManagedAppInstanceUserdataResource{} },
		func() resource.Resource { return &edge.EdgeResource{} },
		func() resource.Resource { return &function.ApiAuthenticatorFunctionResource{} },
		func() resource.Resource { return &function.BitmapperFunctionResource{} },
		func() resource.Resource { return &function.ProcessorFunctionResource{} },
		func() resource.Resource { return &kmskey.KmsKeyResource{} },
		func() resource.Resource { return &managed_node_type.ManagedNodeTypeResource{} },
		func() resource.Resource { return &message_type.MessageTypeResource{} },
		func() resource.Resource { return &node.BitmapRouterNodeResource{} },
		func() resource.Resource { return &node.CrossTenantReceivingNodeResource{} },
		func() resource.Resource { return &node.CrossTenantSendingNodeResource{} },
		func() resource.Resource { return &node.ExternalNodeResource{} },
		func() resource.Resource { return &node.FilesDotComWebhookNodeResource{} },
		func() resource.Resource { return &node.LoadBalancerNodeResource{} },
		func() resource.Resource { return &node.ManagedNodeResource{} },
		func() resource.Resource { return &node.ProcessorNodeResource{} },
		func() resource.Resource { return &node.TimerNodeResource{} },
		func() resource.Resource { return &node.WebhookNodeResource{} },
		func() resource.Resource { return &tenant.TenantResource{} },
		func() resource.Resource { return &user.ApiUserResource{} },
		func() resource.Resource { return &user.TenantUserResource{} },
	}
}
