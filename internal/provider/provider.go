package provider

import (
	"context"
	"os"
	"strings"

	"github.com/credibledata/terraform-provider-credible/internal/client"
	"github.com/credibledata/terraform-provider-credible/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &CredibleProvider{}

// CredibleProvider implements the Credible Terraform provider.
type CredibleProvider struct {
	version string
}

// CredibleProviderModel describes the provider configuration.
type CredibleProviderModel struct {
	URL          types.String `tfsdk:"url"`
	Organization types.String `tfsdk:"organization"`
	APIKey       types.String `tfsdk:"api_key"`
	BearerToken  types.String `tfsdk:"bearer_token"`
}

func New() provider.Provider {
	return &CredibleProvider{
		version: "0.1.0",
	}
}

func (p *CredibleProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "credible"
	resp.Version = p.version
}

func (p *CredibleProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for the Credible Admin API. Manages organizations, projects, connections, packages, and permissions. Authentication can be provided explicitly via api_key/bearer_token, or automatically read from the Credible CLI config (~/.cred).",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "The URL of the Credible controlplane API. Can also be set with the CREDIBLE_URL environment variable.",
				Optional:    true,
			},
			"organization": schema.StringAttribute{
				Description: "The default organization name. Can also be set with the CREDIBLE_ORGANIZATION environment variable. Individual resources can override this.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "A service account API key (sent as 'Authorization: ApiKey <token>'). Can also be set with the CREDIBLE_API_KEY environment variable. Mutually exclusive with bearer_token. If neither api_key nor bearer_token is set, the provider reads the Credible CLI config (~/.cred).",
				Optional:    true,
				Sensitive:   true,
			},
			"bearer_token": schema.StringAttribute{
				Description: "An OAuth2/Auth0 bearer token (sent as 'Authorization: Bearer <token>'). Can also be set with the CREDIBLE_BEARER_TOKEN environment variable. Mutually exclusive with api_key. If neither api_key nor bearer_token is set, the provider reads the Credible CLI config (~/.cred).",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *CredibleProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config CredibleProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := os.Getenv("CREDIBLE_URL")
	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}
	if url == "" {
		resp.Diagnostics.AddError("Missing URL", "The provider 'url' must be set in the provider configuration or via the CREDIBLE_URL environment variable.")
		return
	}

	organization := os.Getenv("CREDIBLE_ORGANIZATION")
	if !config.Organization.IsNull() {
		organization = config.Organization.ValueString()
	}

	// Resolve auth: api_key or bearer_token (env vars, then config)
	apiKey := strings.TrimSpace(os.Getenv("CREDIBLE_API_KEY"))
	if !config.APIKey.IsNull() {
		apiKey = strings.TrimSpace(config.APIKey.ValueString())
	}

	bearerToken := strings.TrimSpace(os.Getenv("CREDIBLE_BEARER_TOKEN"))
	if !config.BearerToken.IsNull() {
		bearerToken = strings.TrimSpace(config.BearerToken.ValueString())
	}

	var authHeader string
	if apiKey != "" && bearerToken != "" {
		resp.Diagnostics.AddError("Conflicting auth", "Only one of 'api_key' or 'bearer_token' may be set, not both.")
		return
	} else if apiKey != "" {
		authHeader = "ApiKey " + apiKey
	} else if bearerToken != "" {
		authHeader = "Bearer " + bearerToken
	} else {
		// Fallback: read from the Credible CLI config file (~/.cred)
		credAuth, err := readCredConfig()
		if err != nil {
			resp.Diagnostics.AddError(
				"Missing authentication",
				"No 'api_key' or 'bearer_token' provided (via config or environment variables), "+
					"and could not read Credible CLI config: "+err.Error()+
					"\n\nEither set api_key/bearer_token or run 'cred login' first.",
			)
			return
		}
		authHeader = credAuth.AuthHeader
		// Use CLI organization as fallback if not explicitly set
		if organization == "" && credAuth.Organization != "" {
			organization = credAuth.Organization
		}
	}

	c := client.NewClient(url, authHeader, organization)

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *CredibleProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewOrganizationResource,
		resources.NewProjectResource,
		resources.NewConnectionResource,
		resources.NewOrganizationPermissionResource,
		resources.NewProjectPermissionResource,
		resources.NewGroupResource,
		resources.NewGroupMemberResource,
		resources.NewPackageResource,
		resources.NewPackageVersionResource,
	}
}

func (p *CredibleProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
