package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/credibledata/terraform-provider-credible/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &ConnectionResource{}
var _ resource.ResourceWithImportState = &ConnectionResource{}

type ConnectionResource struct {
	client *client.Client
}

type ConnectionResourceModel struct {
	Organization     types.String `tfsdk:"organization"`
	Environment      types.String `tfsdk:"environment"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	IncludeTables    types.List   `tfsdk:"include_tables"`
	ExcludeTables    types.List   `tfsdk:"exclude_tables"`
	ExcludeAllTables types.Bool   `tfsdk:"exclude_all_tables"`
	IndexingStatus   types.String `tfsdk:"indexing_status"`

	Proxy *ProxyModel `tfsdk:"proxy"`

	Postgres   *PostgresModel   `tfsdk:"postgres"`
	Bigquery   *BigqueryModel   `tfsdk:"bigquery"`
	Snowflake  *SnowflakeModel  `tfsdk:"snowflake"`
	Trino      *TrinoModel      `tfsdk:"trino"`
	Databricks *DatabricksModel `tfsdk:"databricks"`
	Mysql      *MysqlModel      `tfsdk:"mysql"`
	Duckdb     *DuckdbModel     `tfsdk:"duckdb"`
	Motherduck *MotherduckModel `tfsdk:"motherduck"`
	Ducklake   *DucklakeModel   `tfsdk:"ducklake"`
	Publisher  *PublisherModel  `tfsdk:"publisher"`
}

type ProxyModel struct {
	Type types.String `tfsdk:"type"`
	Ssh  *SshModel    `tfsdk:"ssh"`
}

type SshModel struct {
	Host           types.String `tfsdk:"host"`
	Port           types.Int64  `tfsdk:"port"`
	Username       types.String `tfsdk:"username"`
	PrivateKey     types.String `tfsdk:"private_key"`
	PrivateKeyPass types.String `tfsdk:"private_key_pass"`
	HostKey        types.String `tfsdk:"host_key"`
}

type PostgresModel struct {
	Host             types.String `tfsdk:"host"`
	Port             types.Int64  `tfsdk:"port"`
	DatabaseName     types.String `tfsdk:"database_name"`
	UserName         types.String `tfsdk:"user_name"`
	Password         types.String `tfsdk:"password"`
	ConnectionString types.String `tfsdk:"connection_string"`
	Sslmode          types.String `tfsdk:"sslmode"`
}

type BigqueryModel struct {
	DefaultProjectId          types.String `tfsdk:"default_project_id"`
	BillingProjectId          types.String `tfsdk:"billing_project_id"`
	Location                  types.String `tfsdk:"location"`
	ServiceAccountKeyJson     types.String `tfsdk:"service_account_key_json"`
	MaximumBytesBilled        types.String `tfsdk:"maximum_bytes_billed"`
	QueryTimeoutMilliseconds  types.String `tfsdk:"query_timeout_milliseconds"`
	ImpersonateServiceAccount types.String `tfsdk:"impersonate_service_account"`
}

type SnowflakeModel struct {
	Account                     types.String `tfsdk:"account"`
	Username                    types.String `tfsdk:"username"`
	Password                    types.String `tfsdk:"password"`
	PrivateKey                  types.String `tfsdk:"private_key"`
	PrivateKeyPass              types.String `tfsdk:"private_key_pass"`
	Warehouse                   types.String `tfsdk:"warehouse"`
	Database                    types.String `tfsdk:"database"`
	Schema                      types.String `tfsdk:"schema"`
	Role                        types.String `tfsdk:"role"`
	ResponseTimeoutMilliseconds types.Int64  `tfsdk:"response_timeout_milliseconds"`
}

type TrinoModel struct {
	Server   types.String `tfsdk:"server"`
	Port     types.Int64  `tfsdk:"port"`
	Catalog  types.String `tfsdk:"catalog"`
	Schema   types.String `tfsdk:"schema"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
	PeakaKey types.String `tfsdk:"peaka_key"`
}

type DatabricksModel struct {
	Host              types.String `tfsdk:"host"`
	Path              types.String `tfsdk:"path"`
	Token             types.String `tfsdk:"token"`
	OauthClientId     types.String `tfsdk:"oauth_client_id"`
	OauthClientSecret types.String `tfsdk:"oauth_client_secret"`
	DefaultCatalog    types.String `tfsdk:"default_catalog"`
	DefaultSchema     types.String `tfsdk:"default_schema"`
	SetupSQL          types.String `tfsdk:"setup_sql"`
}

type MysqlModel struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Database types.String `tfsdk:"database"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
}

type DuckdbModel struct {
	AttachedDatabases []AttachedDatabaseModel `tfsdk:"attached_databases"`
}

type AttachedDatabaseModel struct {
	Name      types.String    `tfsdk:"name"`
	Type      types.String    `tfsdk:"type"`
	Bigquery  *BigqueryModel  `tfsdk:"bigquery"`
	Snowflake *SnowflakeModel `tfsdk:"snowflake"`
	Postgres  *PostgresModel  `tfsdk:"postgres"`
	Gcs       *GcsModel       `tfsdk:"gcs"`
	S3        *S3Model        `tfsdk:"s3"`
	Azure     *AzureModel     `tfsdk:"azure"`
}

type MotherduckModel struct {
	AccessToken types.String `tfsdk:"access_token"`
	Database    types.String `tfsdk:"database"`
}

type PublisherModel struct {
	ConnectionUri types.String `tfsdk:"connection_uri"`
	AccessToken   types.String `tfsdk:"access_token"`
}

type DucklakeModel struct {
	Storage *DucklakeStorageModel `tfsdk:"storage"`
	Catalog *DucklakeCatalogModel `tfsdk:"catalog"`
}

type DucklakeStorageModel struct {
	BucketUrl types.String `tfsdk:"bucket_url"`
	S3        *S3Model     `tfsdk:"s3"`
	Gcs       *GcsModel    `tfsdk:"gcs"`
}

type DucklakeCatalogModel struct {
	Postgres       *PostgresModel `tfsdk:"postgres"`
	MetadataSchema types.String   `tfsdk:"metadata_schema"`
}

type GcsModel struct {
	KeyId  types.String `tfsdk:"key_id"`
	Secret types.String `tfsdk:"secret"`
}

type S3Model struct {
	Provider        types.String `tfsdk:"provider"`
	Chain           types.String `tfsdk:"chain"`
	AccessKeyId     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	Region          types.String `tfsdk:"region"`
	Endpoint        types.String `tfsdk:"endpoint"`
	SessionToken    types.String `tfsdk:"session_token"`
}

type AzureModel struct {
	AuthType     types.String `tfsdk:"auth_type"`
	SasUrl       types.String `tfsdk:"sas_url"`
	TenantId     types.String `tfsdk:"tenant_id"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	AccountName  types.String `tfsdk:"account_name"`
	FileUrl      types.String `tfsdk:"file_url"`
}

func NewConnectionResource() resource.Resource {
	return &ConnectionResource{}
}

func (r *ConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection"
}

func (r *ConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a database connection within a Credible environment.",
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Description: "The organization name. Defaults to the provider's organization.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment": schema.StringAttribute{
				Description: "The environment name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The unique name of the connection.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of database connection.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"postgres", "bigquery", "snowflake", "trino", "databricks",
						"mysql", "duckdb", "motherduck", "ducklake", "publisher",
					),
				},
			},
			"include_tables": schema.ListAttribute{
				Description: "List of tables to include (format: schema.table or schema.*).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"exclude_tables": schema.ListAttribute{
				Description: "List of tables to exclude (format: schema.table or schema.*).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"exclude_all_tables": schema.BoolAttribute{
				Description: "Whether to exclude all tables from indexing.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"indexing_status": schema.StringAttribute{
				Description: "Current indexing status of the connection.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"postgres":  postgresBlock(),
			"bigquery":  bigqueryBlock(),
			"snowflake": snowflakeBlock(),
			"trino": schema.SingleNestedBlock{
				Description: "Trino connection configuration.",
				Attributes: map[string]schema.Attribute{
					"server":    schema.StringAttribute{Optional: true, Description: "Trino server hostname."},
					"port":      schema.Int64Attribute{Optional: true, Description: "Trino server port."},
					"catalog":   schema.StringAttribute{Optional: true, Description: "Catalog name."},
					"schema":    schema.StringAttribute{Optional: true, Description: "Schema name."},
					"user":      schema.StringAttribute{Optional: true, Description: "Username."},
					"password":  schema.StringAttribute{Optional: true, Sensitive: true, Description: "Password."},
					"peaka_key": schema.StringAttribute{Optional: true, Sensitive: true, Description: "Peaka API key, for Peaka-hosted Trino clusters."},
				},
			},
			"databricks": schema.SingleNestedBlock{
				Description: "Databricks SQL warehouse connection configuration.",
				Attributes: map[string]schema.Attribute{
					"host":                schema.StringAttribute{Optional: true, Description: "Workspace host (e.g. dbc-xxxxxxxx-xxxx.cloud.databricks.com)."},
					"path":                schema.StringAttribute{Optional: true, Description: "SQL warehouse HTTP path (e.g. /sql/1.0/warehouses/<warehouse-id>)."},
					"token":               schema.StringAttribute{Optional: true, Sensitive: true, Description: "Personal access token."},
					"oauth_client_id":     schema.StringAttribute{Optional: true, Description: "OAuth M2M client ID (service principal)."},
					"oauth_client_secret": schema.StringAttribute{Optional: true, Sensitive: true, Description: "OAuth M2M client secret (service principal)."},
					"default_catalog":     schema.StringAttribute{Optional: true, Description: "Default Unity Catalog for queries."},
					"default_schema":      schema.StringAttribute{Optional: true, Description: "Default schema for queries."},
					"setup_sql":           schema.StringAttribute{Optional: true, Description: "SQL run when the connection is established."},
				},
			},
			"mysql": schema.SingleNestedBlock{
				Description: "MySQL connection configuration.",
				Attributes: map[string]schema.Attribute{
					"host":     schema.StringAttribute{Optional: true, Description: "MySQL server hostname."},
					"port":     schema.Int64Attribute{Optional: true, Description: "MySQL server port."},
					"database": schema.StringAttribute{Optional: true, Description: "Database name."},
					"user":     schema.StringAttribute{Optional: true, Description: "Username."},
					"password": schema.StringAttribute{Optional: true, Sensitive: true, Description: "Password."},
				},
			},
			"duckdb": schema.SingleNestedBlock{
				Description: "DuckDB connection configuration. Exposes data-source intent only; " +
					"database files, filesystem policy and resource limits are owned by the deployment.",
				Blocks: map[string]schema.Block{
					"attached_databases": schema.ListNestedBlock{
						Description: "Warehouses attached to this DuckDB connection.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{Optional: true, Description: "Attachment name (must be a plain identifier)."},
								"type": schema.StringAttribute{
									Optional:    true,
									Description: "Attached database type.",
									Validators: []validator.String{
										stringvalidator.OneOf("bigquery", "snowflake", "postgres", "gcs", "s3", "azure"),
									},
								},
							},
							Blocks: map[string]schema.Block{
								"bigquery":  bigqueryBlock(),
								"snowflake": snowflakeBlock(),
								"postgres":  postgresBlock(),
								"gcs":       gcsBlock(),
								"s3":        s3Block(),
								"azure":     azureBlock(),
							},
						},
					},
				},
			},
			"motherduck": schema.SingleNestedBlock{
				Description: "MotherDuck connection configuration.",
				Attributes: map[string]schema.Attribute{
					"access_token": schema.StringAttribute{Optional: true, Sensitive: true, Description: "MotherDuck access token."},
					"database":     schema.StringAttribute{Optional: true, Description: "MotherDuck database name."},
				},
			},
			"ducklake": schema.SingleNestedBlock{
				Description: "DuckLake lakehouse connection configuration.",
				Blocks: map[string]schema.Block{
					"storage": schema.SingleNestedBlock{
						Description: "Data storage configuration. `bucket_url` is required when this block is set.",
						Attributes: map[string]schema.Attribute{
							"bucket_url": schema.StringAttribute{Optional: true, Description: "Storage bucket URL (e.g. s3://my-bucket/path or gs://my-bucket/path)."},
						},
						Blocks: map[string]schema.Block{
							"s3":  s3Block(),
							"gcs": gcsBlock(),
						},
					},
					"catalog": schema.SingleNestedBlock{
						Description: "Catalog metadata configuration. `postgres` is required when this block is set.",
						Attributes: map[string]schema.Attribute{
							"metadata_schema": schema.StringAttribute{
								Optional: true,
								Description: "Schema holding this DuckLake's metadata tables. Lets several catalogs " +
									"share one catalog database. An organizational boundary, not an access-control one.",
							},
						},
						Blocks: map[string]schema.Block{
							"postgres": postgresBlock(),
						},
					},
				},
			},
			"publisher": schema.SingleNestedBlock{
				Description: "Malloy Publisher proxy connection. Proxies SQL to a remote Publisher dataplane.",
				Attributes: map[string]schema.Attribute{
					"connection_uri": schema.StringAttribute{Optional: true, Description: "Full URI of the remote connection. Required when this block is set."},
					"access_token":   schema.StringAttribute{Optional: true, Sensitive: true, Description: "Bearer token for the remote dataplane."},
				},
			},
			"proxy": schema.SingleNestedBlock{
				Description: "Optional network proxy for reaching a database that is not directly routable.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional:    true,
						Description: "Proxy mechanism.",
						Validators: []validator.String{
							stringvalidator.OneOf("ssh"),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"ssh": schema.SingleNestedBlock{
						Description: "SSH bastion configuration. Authentication is public-key only.",
						Attributes: map[string]schema.Attribute{
							"host":     schema.StringAttribute{Optional: true, Description: "Bastion hostname or IP address."},
							"port":     schema.Int64Attribute{Optional: true, Description: "Bastion SSH port. The API defaults this to 22."},
							"username": schema.StringAttribute{Optional: true, Description: "SSH username on the bastion."},
							"private_key": schema.StringAttribute{
								Optional: true, Sensitive: true,
								Description: "PEM-encoded SSH private key. Leave blank when updating to keep the stored key.",
							},
							"private_key_pass": schema.StringAttribute{
								Optional: true, Sensitive: true,
								Description: "Passphrase for the private key. Leave blank when updating to keep the stored one.",
							},
							"host_key": schema.StringAttribute{
								Optional: true,
								Description: "Pinned bastion host public key(s) as OpenSSH known_hosts lines. " +
									"Omitted means the tunnel connects without host-key verification.",
							},
						},
					},
				},
			},
		},
	}
}

// The storage/warehouse blocks below are reused in several places (a DuckDB
// attached database, DuckLake storage, DuckLake catalog), so they are built by
// helpers rather than duplicated -- keeping one definition per concept.

func postgresBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "PostgreSQL connection configuration.",
		Attributes: map[string]schema.Attribute{
			"host":              schema.StringAttribute{Optional: true, Description: "PostgreSQL server hostname."},
			"port":              schema.Int64Attribute{Optional: true, Description: "PostgreSQL server port."},
			"database_name":     schema.StringAttribute{Optional: true, Description: "Database name."},
			"user_name":         schema.StringAttribute{Optional: true, Description: "Username."},
			"password":          schema.StringAttribute{Optional: true, Sensitive: true, Description: "Password."},
			"connection_string": schema.StringAttribute{Optional: true, Sensitive: true, Description: "Full connection string (alternative to individual params)."},
			"sslmode": schema.StringAttribute{
				Optional:    true,
				Description: "TLS mode. Only valid on a proxied connection; a direct connection rejects this field.",
				Validators: []validator.String{
					stringvalidator.OneOf("disable", "no-verify", "verify-ca", "verify-full"),
				},
			},
		},
	}
}

func bigqueryBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "BigQuery connection configuration.",
		Attributes: map[string]schema.Attribute{
			"default_project_id":         schema.StringAttribute{Optional: true, Description: "Default BigQuery project ID."},
			"billing_project_id":         schema.StringAttribute{Optional: true, Description: "Billing project ID."},
			"location":                   schema.StringAttribute{Optional: true, Description: "Dataset location."},
			"service_account_key_json":   schema.StringAttribute{Optional: true, Sensitive: true, Description: "Service account key JSON."},
			"maximum_bytes_billed":       schema.StringAttribute{Optional: true, Description: "Maximum bytes billed."},
			"query_timeout_milliseconds": schema.StringAttribute{Optional: true, Description: "Query timeout in milliseconds."},
			"impersonate_service_account": schema.StringAttribute{
				Optional:    true,
				Description: "Service account to impersonate. Mutually exclusive with service_account_key_json.",
			},
		},
	}
}

func snowflakeBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Snowflake connection configuration.",
		Attributes: map[string]schema.Attribute{
			"account":                       schema.StringAttribute{Optional: true, Description: "Snowflake account identifier."},
			"username":                      schema.StringAttribute{Optional: true, Description: "Username."},
			"password":                      schema.StringAttribute{Optional: true, Sensitive: true, Description: "Password."},
			"private_key":                   schema.StringAttribute{Optional: true, Sensitive: true, Description: "Private key for authentication."},
			"private_key_pass":              schema.StringAttribute{Optional: true, Sensitive: true, Description: "Private key passphrase."},
			"warehouse":                     schema.StringAttribute{Optional: true, Description: "Warehouse name."},
			"database":                      schema.StringAttribute{Optional: true, Description: "Database name."},
			"schema":                        schema.StringAttribute{Optional: true, Description: "Schema name."},
			"role":                          schema.StringAttribute{Optional: true, Description: "Role name."},
			"response_timeout_milliseconds": schema.Int64Attribute{Optional: true, Description: "Response timeout in milliseconds."},
		},
	}
}

func gcsBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Google Cloud Storage configuration. Both attributes are required when this block is set.",
		Attributes: map[string]schema.Attribute{
			"key_id": schema.StringAttribute{Optional: true, Description: "GCS HMAC access key ID."},
			"secret": schema.StringAttribute{Optional: true, Sensitive: true, Description: "GCS HMAC secret key."},
		},
	}
}

func s3Block() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "AWS S3 configuration.",
		Attributes: map[string]schema.Attribute{
			"provider": schema.StringAttribute{
				Optional: true,
				Description: "How credentials are obtained; the API defaults to `config` when omitted. " +
					"`config` requires access_key_id and secret_access_key; `credential_chain` resolves " +
					"them from the host and rejects a supplied key.",
				Validators: []validator.String{
					stringvalidator.OneOf("config", "credential_chain"),
				},
			},
			"chain":             schema.StringAttribute{Optional: true, Description: "Semicolon-separated credential providers, for `credential_chain`."},
			"access_key_id":     schema.StringAttribute{Optional: true, Description: "AWS access key ID."},
			"secret_access_key": schema.StringAttribute{Optional: true, Sensitive: true, Description: "AWS secret access key."},
			"region":            schema.StringAttribute{Optional: true, Description: "AWS region. The API defaults to us-east-1."},
			"endpoint":          schema.StringAttribute{Optional: true, Description: "Custom S3-compatible endpoint URL (e.g. MinIO)."},
			"session_token":     schema.StringAttribute{Optional: true, Sensitive: true, Description: "AWS session token for temporary credentials."},
		},
	}
}

func azureBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Azure Data Lake / Blob Storage configuration. `auth_type` is required when this block is set.",
		Attributes: map[string]schema.Attribute{
			"auth_type": schema.StringAttribute{
				Optional:    true,
				Description: "Authentication method.",
				Validators: []validator.String{
					stringvalidator.OneOf("service_principal", "sas_token"),
				},
			},
			"sas_url":       schema.StringAttribute{Optional: true, Sensitive: true, Description: "Full SAS URL including token; required for sas_token auth."},
			"tenant_id":     schema.StringAttribute{Optional: true, Description: "Azure AD tenant ID; required for service_principal."},
			"client_id":     schema.StringAttribute{Optional: true, Description: "Azure AD application (client) ID; required for service_principal."},
			"client_secret": schema.StringAttribute{Optional: true, Sensitive: true, Description: "Azure AD client secret; required for service_principal."},
			"account_name":  schema.StringAttribute{Optional: true, Description: "Storage account name; required for service_principal."},
			"file_url":      schema.StringAttribute{Optional: true, Description: "Azure file URL to query; required for service_principal."},
		},
	}
}

func (r *ConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *ConnectionResource) getOrg(model *ConnectionResourceModel) string {
	if !model.Organization.IsNull() && !model.Organization.IsUnknown() {
		return model.Organization.ValueString()
	}
	return r.client.Organization
}

func (r *ConnectionResource) modelToAPI(ctx context.Context, model *ConnectionResourceModel) *client.Connection {
	conn := &client.Connection{
		Name: model.Name.ValueString(),
		Type: model.Type.ValueString(),
	}

	// Include/Exclude tables
	if !model.IncludeTables.IsNull() {
		var tables []string
		model.IncludeTables.ElementsAs(ctx, &tables, false)
		conn.IncludeTables = tables
	}
	if !model.ExcludeTables.IsNull() {
		var tables []string
		model.ExcludeTables.ElementsAs(ctx, &tables, false)
		conn.ExcludeTables = tables
	}
	if !model.ExcludeAllTables.IsNull() && !model.ExcludeAllTables.IsUnknown() {
		v := model.ExcludeAllTables.ValueBool()
		conn.ExcludeAllTables = &v
	}

	// Type-specific configuration
	// Each engine maps through the same helper used by the nested (DuckDB attached
	// database, DuckLake) positions, so a field added to one is not missed in the other.
	conn.PostgresConnection = postgresToAPI(model.Postgres)
	conn.BigqueryConnection = bigqueryToAPI(model.Bigquery)
	conn.SnowflakeConnection = snowflakeToAPI(model.Snowflake)

	if model.Trino != nil {
		conn.TrinoConnection = &client.TrinoConnection{
			Server:   str(model.Trino.Server),
			Port:     intPtr(model.Trino.Port),
			Catalog:  str(model.Trino.Catalog),
			Schema:   str(model.Trino.Schema),
			User:     str(model.Trino.User),
			Password: str(model.Trino.Password),
			PeakaKey: str(model.Trino.PeakaKey),
		}
	}

	if model.Mysql != nil {
		conn.MysqlConnection = &client.MysqlConnection{
			Host:     str(model.Mysql.Host),
			Port:     intPtr(model.Mysql.Port),
			Database: str(model.Mysql.Database),
			User:     str(model.Mysql.User),
			Password: str(model.Mysql.Password),
		}
	}

	if model.Databricks != nil {
		conn.DatabricksConnection = &client.DatabricksConnection{
			Host:              str(model.Databricks.Host),
			Path:              str(model.Databricks.Path),
			Token:             str(model.Databricks.Token),
			OauthClientId:     str(model.Databricks.OauthClientId),
			OauthClientSecret: str(model.Databricks.OauthClientSecret),
			DefaultCatalog:    str(model.Databricks.DefaultCatalog),
			DefaultSchema:     str(model.Databricks.DefaultSchema),
			SetupSQL:          str(model.Databricks.SetupSQL),
		}
	}

	if model.Duckdb != nil {
		conn.DuckdbConnection = &client.DuckdbConnection{}
		for _, ad := range model.Duckdb.AttachedDatabases {
			conn.DuckdbConnection.AttachedDatabases = append(
				conn.DuckdbConnection.AttachedDatabases,
				client.AttachedDatabase{
					Name:                str(ad.Name),
					Type:                str(ad.Type),
					BigqueryConnection:  bigqueryToAPI(ad.Bigquery),
					SnowflakeConnection: snowflakeToAPI(ad.Snowflake),
					PostgresConnection:  postgresToAPI(ad.Postgres),
					GcsConnection:       gcsToAPI(ad.Gcs),
					S3Connection:        s3ToAPI(ad.S3),
					AzureConnection:     azureToAPI(ad.Azure),
				},
			)
		}
	}

	if model.Motherduck != nil {
		conn.MotherDuckConnection = &client.MotherDuckConnection{
			AccessToken: str(model.Motherduck.AccessToken),
			Database:    str(model.Motherduck.Database),
		}
	}

	if model.Publisher != nil {
		conn.PublisherConnection = &client.PublisherConnection{
			ConnectionUri: str(model.Publisher.ConnectionUri),
			AccessToken:   str(model.Publisher.AccessToken),
		}
	}

	if model.Ducklake != nil {
		conn.DucklakeConnection = &client.DucklakeConnection{}
		if s := model.Ducklake.Storage; s != nil {
			conn.DucklakeConnection.Storage = &client.DucklakeStorage{
				BucketUrl:     str(s.BucketUrl),
				S3Connection:  s3ToAPI(s.S3),
				GcsConnection: gcsToAPI(s.Gcs),
			}
		}
		if c := model.Ducklake.Catalog; c != nil {
			conn.DucklakeConnection.Catalog = &client.DucklakeCatalog{
				PostgresConnection: postgresToAPI(c.Postgres),
				MetadataSchema:     str(c.MetadataSchema),
			}
		}
	}

	if model.Proxy != nil {
		conn.Proxy = &client.ConnectionProxy{Type: str(model.Proxy.Type)}
		if s := model.Proxy.Ssh; s != nil {
			conn.Proxy.Ssh = &client.SshProxyConfig{
				Host:           str(s.Host),
				Port:           intPtr(s.Port),
				Username:       str(s.Username),
				PrivateKey:     str(s.PrivateKey),
				PrivateKeyPass: str(s.PrivateKeyPass),
				HostKey:        str(s.HostKey),
			}
		}
	}

	return conn
}

// stringListOrNull keeps an absent list null rather than an empty list: the two are
// distinct to Terraform, and mapping absent to [] would plan a permanent diff
// against a config that omits the attribute.
func stringListOrNull(ctx context.Context, values []string) types.List {
	if values == nil {
		return types.ListNull(types.StringType)
	}
	list, diags := types.ListValueFrom(ctx, types.StringType, values)
	if diags.HasError() {
		return types.ListNull(types.StringType)
	}
	return list
}

// str and intPtr unwrap a Terraform value, mapping null/unknown to the Go zero
// value so `omitempty` drops the field rather than sending an empty one.
func str(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

func intPtr(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := int(v.ValueInt64())
	return &i
}

func postgresToAPI(m *PostgresModel) *client.PostgresConnection {
	if m == nil {
		return nil
	}
	return &client.PostgresConnection{
		Host:             str(m.Host),
		Port:             intPtr(m.Port),
		DatabaseName:     str(m.DatabaseName),
		UserName:         str(m.UserName),
		Password:         str(m.Password),
		ConnectionString: str(m.ConnectionString),
		Sslmode:          str(m.Sslmode),
	}
}

func bigqueryToAPI(m *BigqueryModel) *client.BigqueryConnection {
	if m == nil {
		return nil
	}
	return &client.BigqueryConnection{
		DefaultProjectId:          str(m.DefaultProjectId),
		BillingProjectId:          str(m.BillingProjectId),
		Location:                  str(m.Location),
		ServiceAccountKeyJson:     str(m.ServiceAccountKeyJson),
		MaximumBytesBilled:        str(m.MaximumBytesBilled),
		QueryTimeoutMilliseconds:  str(m.QueryTimeoutMilliseconds),
		ImpersonateServiceAccount: str(m.ImpersonateServiceAccount),
	}
}

func snowflakeToAPI(m *SnowflakeModel) *client.SnowflakeConnection {
	if m == nil {
		return nil
	}
	return &client.SnowflakeConnection{
		Account:                     str(m.Account),
		Username:                    str(m.Username),
		Password:                    str(m.Password),
		PrivateKey:                  str(m.PrivateKey),
		PrivateKeyPass:              str(m.PrivateKeyPass),
		Warehouse:                   str(m.Warehouse),
		Database:                    str(m.Database),
		Schema:                      str(m.Schema),
		Role:                        str(m.Role),
		ResponseTimeoutMilliseconds: intPtr(m.ResponseTimeoutMilliseconds),
	}
}

func gcsToAPI(m *GcsModel) *client.GCSConnection {
	if m == nil {
		return nil
	}
	return &client.GCSConnection{KeyId: str(m.KeyId), Secret: str(m.Secret)}
}

func s3ToAPI(m *S3Model) *client.S3Connection {
	if m == nil {
		return nil
	}
	return &client.S3Connection{
		Provider:        str(m.Provider),
		Chain:           str(m.Chain),
		AccessKeyId:     str(m.AccessKeyId),
		SecretAccessKey: str(m.SecretAccessKey),
		Region:          str(m.Region),
		Endpoint:        str(m.Endpoint),
		SessionToken:    str(m.SessionToken),
	}
}

func azureToAPI(m *AzureModel) *client.AzureConnection {
	if m == nil {
		return nil
	}
	return &client.AzureConnection{
		AuthType:     str(m.AuthType),
		SasUrl:       str(m.SasUrl),
		TenantId:     str(m.TenantId),
		ClientId:     str(m.ClientId),
		ClientSecret: str(m.ClientSecret),
		AccountName:  str(m.AccountName),
		FileUrl:      str(m.FileUrl),
	}
}

// apiToModel updates the model with values from API response.
// Sensitive fields (passwords, keys) are preserved from the plan since the API won't return them.
func (r *ConnectionResource) apiToModel(ctx context.Context, result *client.Connection, model *ConnectionResourceModel, preserveSensitive *ConnectionResourceModel) {
	model.Name = types.StringValue(result.Name)
	model.Type = types.StringValue(result.Type)
	model.IndexingStatus = types.StringValue(result.IndexingStatus)

	if result.ExcludeAllTables != nil {
		model.ExcludeAllTables = types.BoolValue(*result.ExcludeAllTables)
	}

	// Table filters are non-sensitive, so the API does return them: refresh from the
	// response rather than re-asserting prior state, otherwise an out-of-band edit to
	// the filters never shows as drift.
	model.IncludeTables = stringListOrNull(ctx, result.IncludeTables)
	model.ExcludeTables = stringListOrNull(ctx, result.ExcludeTables)

	// Rebuild every engine block from the response. Refreshing in place would leave
	// blocks the response no longer carries (so an engine changed out of band yields
	// a state no config could produce), and would leave non-secret fields showing
	// their stale configured value instead of drift. Secrets are restored from prior
	// below, since the API never returns them.
	//
	// prior must be read AFTER this point only: callers legitimately pass the same
	// model as both arguments, so the reads below take their values from the snapshot
	// captured here rather than from the model being overwritten.
	prior := snapshotSecrets(model, preserveSensitive)
	clearEngineBlocks(model)
	populateBlocksFromAPI(result, model)
	preserveSensitive = prior

	// For sensitive fields, preserve what was in the plan/state since API doesn't return them
	if preserveSensitive != nil && preserveSensitive.Postgres != nil && model.Postgres != nil {
		if !preserveSensitive.Postgres.Password.IsNull() {
			model.Postgres.Password = preserveSensitive.Postgres.Password
		}
		if !preserveSensitive.Postgres.ConnectionString.IsNull() {
			model.Postgres.ConnectionString = preserveSensitive.Postgres.ConnectionString
		}
	}
	if preserveSensitive != nil && preserveSensitive.Bigquery != nil && model.Bigquery != nil {
		if !preserveSensitive.Bigquery.ServiceAccountKeyJson.IsNull() {
			model.Bigquery.ServiceAccountKeyJson = preserveSensitive.Bigquery.ServiceAccountKeyJson
		}
	}
	if preserveSensitive != nil && preserveSensitive.Snowflake != nil && model.Snowflake != nil {
		if !preserveSensitive.Snowflake.Password.IsNull() {
			model.Snowflake.Password = preserveSensitive.Snowflake.Password
		}
		if !preserveSensitive.Snowflake.PrivateKey.IsNull() {
			model.Snowflake.PrivateKey = preserveSensitive.Snowflake.PrivateKey
		}
		if !preserveSensitive.Snowflake.PrivateKeyPass.IsNull() {
			model.Snowflake.PrivateKeyPass = preserveSensitive.Snowflake.PrivateKeyPass
		}
	}
	if preserveSensitive != nil && preserveSensitive.Mysql != nil && model.Mysql != nil {
		if !preserveSensitive.Mysql.Password.IsNull() {
			model.Mysql.Password = preserveSensitive.Mysql.Password
		}
	}
	if preserveSensitive != nil && preserveSensitive.Trino != nil && model.Trino != nil {
		if !preserveSensitive.Trino.Password.IsNull() {
			model.Trino.Password = preserveSensitive.Trino.Password
		}
		if !preserveSensitive.Trino.PeakaKey.IsNull() {
			model.Trino.PeakaKey = preserveSensitive.Trino.PeakaKey
		}
	}
	if preserveSensitive != nil && preserveSensitive.Databricks != nil && model.Databricks != nil {
		if !preserveSensitive.Databricks.Token.IsNull() {
			model.Databricks.Token = preserveSensitive.Databricks.Token
		}
		if !preserveSensitive.Databricks.OauthClientSecret.IsNull() {
			model.Databricks.OauthClientSecret = preserveSensitive.Databricks.OauthClientSecret
		}
	}
	if preserveSensitive != nil && preserveSensitive.Motherduck != nil && model.Motherduck != nil {
		if !preserveSensitive.Motherduck.AccessToken.IsNull() {
			model.Motherduck.AccessToken = preserveSensitive.Motherduck.AccessToken
		}
	}
	if preserveSensitive != nil && preserveSensitive.Publisher != nil && model.Publisher != nil {
		if !preserveSensitive.Publisher.AccessToken.IsNull() {
			model.Publisher.AccessToken = preserveSensitive.Publisher.AccessToken
		}
	}
	if preserveSensitive != nil && preserveSensitive.Duckdb != nil && model.Duckdb != nil {
		// Pair by name, not by position. `name` is what identifies an attachment;
		// removing one from config shifts every later index, and index pairing would
		// then restore each remaining attachment the neighbouring database's
		// credential -- writing the wrong secret to the wrong warehouse.
		priorByName := make(map[string]*AttachedDatabaseModel, len(preserveSensitive.Duckdb.AttachedDatabases))
		for i := range preserveSensitive.Duckdb.AttachedDatabases {
			p := &preserveSensitive.Duckdb.AttachedDatabases[i]
			if name := p.Name.ValueString(); name != "" {
				priorByName[name] = p
			}
		}
		for i := range model.Duckdb.AttachedDatabases {
			current := &model.Duckdb.AttachedDatabases[i]
			// An unnamed or newly added attachment has no prior secrets to restore.
			if p, ok := priorByName[current.Name.ValueString()]; ok {
				preserveAttachedSecrets(current, p)
			}
		}
	}
	if preserveSensitive != nil && preserveSensitive.Ducklake != nil && model.Ducklake != nil {
		if s, ps := model.Ducklake.Storage, preserveSensitive.Ducklake.Storage; s != nil && ps != nil {
			preserveS3Secrets(s.S3, ps.S3)
			preserveGcsSecrets(s.Gcs, ps.Gcs)
		}
		if c, pc := model.Ducklake.Catalog, preserveSensitive.Ducklake.Catalog; c != nil && pc != nil {
			preservePostgresSecrets(c.Postgres, pc.Postgres)
		}
	}
	if preserveSensitive != nil && preserveSensitive.Proxy != nil && model.Proxy != nil {
		if s, ps := model.Proxy.Ssh, preserveSensitive.Proxy.Ssh; s != nil && ps != nil {
			if !ps.PrivateKey.IsNull() {
				s.PrivateKey = ps.PrivateKey
			}
			if !ps.PrivateKeyPass.IsNull() {
				s.PrivateKeyPass = ps.PrivateKeyPass
			}
		}
	}
}

// The API never returns secrets, so each helper below carries the configured
// value forward. Without this the attribute reads back null and Terraform plans a
// permanent diff on every run.

func preserveAttachedSecrets(model, prior *AttachedDatabaseModel) {
	preservePostgresSecrets(model.Postgres, prior.Postgres)
	preserveS3Secrets(model.S3, prior.S3)
	preserveGcsSecrets(model.Gcs, prior.Gcs)
	if model.Bigquery != nil && prior.Bigquery != nil && !prior.Bigquery.ServiceAccountKeyJson.IsNull() {
		model.Bigquery.ServiceAccountKeyJson = prior.Bigquery.ServiceAccountKeyJson
	}
	if model.Snowflake != nil && prior.Snowflake != nil {
		if !prior.Snowflake.Password.IsNull() {
			model.Snowflake.Password = prior.Snowflake.Password
		}
		if !prior.Snowflake.PrivateKey.IsNull() {
			model.Snowflake.PrivateKey = prior.Snowflake.PrivateKey
		}
		if !prior.Snowflake.PrivateKeyPass.IsNull() {
			model.Snowflake.PrivateKeyPass = prior.Snowflake.PrivateKeyPass
		}
	}
	if model.Azure != nil && prior.Azure != nil {
		if !prior.Azure.ClientSecret.IsNull() {
			model.Azure.ClientSecret = prior.Azure.ClientSecret
		}
		if !prior.Azure.SasUrl.IsNull() {
			model.Azure.SasUrl = prior.Azure.SasUrl
		}
	}
}

func preservePostgresSecrets(model, prior *PostgresModel) {
	if model == nil || prior == nil {
		return
	}
	if !prior.Password.IsNull() {
		model.Password = prior.Password
	}
	if !prior.ConnectionString.IsNull() {
		model.ConnectionString = prior.ConnectionString
	}
}

func preserveS3Secrets(model, prior *S3Model) {
	if model == nil || prior == nil {
		return
	}
	if !prior.SecretAccessKey.IsNull() {
		model.SecretAccessKey = prior.SecretAccessKey
	}
	if !prior.SessionToken.IsNull() {
		model.SessionToken = prior.SessionToken
	}
}

func preserveGcsSecrets(model, prior *GcsModel) {
	if model == nil || prior == nil {
		return
	}
	if !prior.Secret.IsNull() {
		model.Secret = prior.Secret
	}
}

func (r *ConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	if org == "" {
		resp.Diagnostics.AddError("Missing organization", "Organization must be set either on the resource or provider.")
		return
	}

	conn := r.modelToAPI(ctx, &plan)

	tflog.Debug(ctx, "Creating connection", map[string]interface{}{"org": org, "environment": plan.Environment.ValueString(), "name": conn.Name})

	_, err := r.client.CreateConnection(org, plan.Environment.ValueString(), conn)
	if err != nil {
		resp.Diagnostics.AddError("Error creating connection", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)

	// Read back to get computed fields
	result, err := r.client.GetConnection(org, plan.Environment.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading connection after create", err.Error())
		return
	}

	r.apiToModel(ctx, result, &plan, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	result, err := r.client.GetConnection(org, state.Environment.ValueString(), state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading connection", err.Error())
		return
	}

	preserveSensitive := state
	state.Organization = types.StringValue(org)
	r.apiToModel(ctx, result, &state, &preserveSensitive)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	conn := r.modelToAPI(ctx, &plan)

	_, err := r.client.UpdateConnection(org, plan.Environment.ValueString(), plan.Name.ValueString(), conn)
	if err != nil {
		resp.Diagnostics.AddError("Error updating connection", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)

	result, err := r.client.GetConnection(org, plan.Environment.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading connection after update", err.Error())
		return
	}

	r.apiToModel(ctx, result, &plan, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	err := r.client.DeleteConnection(org, state.Environment.ValueString(), state.Name.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting connection", err.Error())
	}
}

func (r *ConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be in the format: organization/environment/connection")
		return
	}

	org, environment, name := parts[0], parts[1], parts[2]
	result, err := r.client.GetConnection(org, environment, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing connection", err.Error())
		return
	}

	state := ConnectionResourceModel{
		Organization: types.StringValue(org),
		Environment:  types.StringValue(environment),
	}
	// apiToModel rebuilds the engine block from the response, so import needs no
	// extra work. Secrets are absent (the API never returns them) and must come from
	// config -- the acceptance tests cover that via ImportStateVerifyIgnore.
	r.apiToModel(ctx, result, &state, nil)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// populateBlocksFromAPI reconstructs the engine and proxy blocks from an API
// response. Used on import, where there is no prior state to refresh into.
func populateBlocksFromAPI(result *client.Connection, model *ConnectionResourceModel) {
	if c := result.PostgresConnection; c != nil {
		model.Postgres = postgresFromAPI(c)
	}
	if c := result.BigqueryConnection; c != nil {
		model.Bigquery = bigqueryFromAPI(c)
	}
	if c := result.SnowflakeConnection; c != nil {
		model.Snowflake = snowflakeFromAPI(c)
	}
	if c := result.TrinoConnection; c != nil {
		model.Trino = &TrinoModel{
			Server:   optString(c.Server),
			Port:     optInt(c.Port),
			Catalog:  optString(c.Catalog),
			Schema:   optString(c.Schema),
			User:     optString(c.User),
			Password: optString(c.Password),
			PeakaKey: optString(c.PeakaKey),
		}
	}
	if c := result.DatabricksConnection; c != nil {
		model.Databricks = &DatabricksModel{
			Host:              optString(c.Host),
			Path:              optString(c.Path),
			Token:             optString(c.Token),
			OauthClientId:     optString(c.OauthClientId),
			OauthClientSecret: optString(c.OauthClientSecret),
			DefaultCatalog:    optString(c.DefaultCatalog),
			DefaultSchema:     optString(c.DefaultSchema),
			SetupSQL:          optString(c.SetupSQL),
		}
	}
	if c := result.MysqlConnection; c != nil {
		model.Mysql = &MysqlModel{
			Host:     optString(c.Host),
			Port:     optInt(c.Port),
			Database: optString(c.Database),
			User:     optString(c.User),
			Password: optString(c.Password),
		}
	}
	if c := result.DuckdbConnection; c != nil {
		model.Duckdb = &DuckdbModel{}
		for _, ad := range c.AttachedDatabases {
			model.Duckdb.AttachedDatabases = append(model.Duckdb.AttachedDatabases, AttachedDatabaseModel{
				Name:      optString(ad.Name),
				Type:      optString(ad.Type),
				Bigquery:  bigqueryFromAPI(ad.BigqueryConnection),
				Snowflake: snowflakeFromAPI(ad.SnowflakeConnection),
				Postgres:  postgresFromAPI(ad.PostgresConnection),
				Gcs:       gcsFromAPI(ad.GcsConnection),
				S3:        s3FromAPI(ad.S3Connection),
				Azure:     azureFromAPI(ad.AzureConnection),
			})
		}
	}
	if c := result.MotherDuckConnection; c != nil {
		model.Motherduck = &MotherduckModel{
			AccessToken: optString(c.AccessToken),
			Database:    optString(c.Database),
		}
	}
	if c := result.PublisherConnection; c != nil {
		model.Publisher = &PublisherModel{
			ConnectionUri: optString(c.ConnectionUri),
			AccessToken:   optString(c.AccessToken),
		}
	}
	if c := result.DucklakeConnection; c != nil {
		model.Ducklake = &DucklakeModel{}
		if s := c.Storage; s != nil {
			model.Ducklake.Storage = &DucklakeStorageModel{
				BucketUrl: optString(s.BucketUrl),
				S3:        s3FromAPI(s.S3Connection),
				Gcs:       gcsFromAPI(s.GcsConnection),
			}
		}
		if cat := c.Catalog; cat != nil {
			model.Ducklake.Catalog = &DucklakeCatalogModel{
				Postgres:       postgresFromAPI(cat.PostgresConnection),
				MetadataSchema: optString(cat.MetadataSchema),
			}
		}
	}
	if p := result.Proxy; p != nil {
		model.Proxy = &ProxyModel{Type: optString(p.Type)}
		if s := p.Ssh; s != nil {
			model.Proxy.Ssh = &SshModel{
				Host:           optString(s.Host),
				Port:           optInt(s.Port),
				Username:       optString(s.Username),
				PrivateKey:     optString(s.PrivateKey),
				PrivateKeyPass: optString(s.PrivateKeyPass),
				HostKey:        optString(s.HostKey),
			}
		}
	}
}

// optString maps the API's empty string to a null attribute: the wire format uses
// `omitempty`, so "" means absent rather than an explicitly empty value.
func optString(v string) types.String {
	if v == "" {
		return types.StringNull()
	}
	return types.StringValue(v)
}

func optInt(v *int) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*v))
}

func postgresFromAPI(c *client.PostgresConnection) *PostgresModel {
	if c == nil {
		return nil
	}
	return &PostgresModel{
		Host:             optString(c.Host),
		Port:             optInt(c.Port),
		DatabaseName:     optString(c.DatabaseName),
		UserName:         optString(c.UserName),
		Password:         optString(c.Password),
		ConnectionString: optString(c.ConnectionString),
		Sslmode:          optString(c.Sslmode),
	}
}

func bigqueryFromAPI(c *client.BigqueryConnection) *BigqueryModel {
	if c == nil {
		return nil
	}
	return &BigqueryModel{
		DefaultProjectId:          optString(c.DefaultProjectId),
		BillingProjectId:          optString(c.BillingProjectId),
		Location:                  optString(c.Location),
		ServiceAccountKeyJson:     optString(c.ServiceAccountKeyJson),
		MaximumBytesBilled:        optString(c.MaximumBytesBilled),
		QueryTimeoutMilliseconds:  optString(c.QueryTimeoutMilliseconds),
		ImpersonateServiceAccount: optString(c.ImpersonateServiceAccount),
	}
}

func snowflakeFromAPI(c *client.SnowflakeConnection) *SnowflakeModel {
	if c == nil {
		return nil
	}
	return &SnowflakeModel{
		Account:                     optString(c.Account),
		Username:                    optString(c.Username),
		Password:                    optString(c.Password),
		PrivateKey:                  optString(c.PrivateKey),
		PrivateKeyPass:              optString(c.PrivateKeyPass),
		Warehouse:                   optString(c.Warehouse),
		Database:                    optString(c.Database),
		Schema:                      optString(c.Schema),
		Role:                        optString(c.Role),
		ResponseTimeoutMilliseconds: optInt(c.ResponseTimeoutMilliseconds),
	}
}

func gcsFromAPI(c *client.GCSConnection) *GcsModel {
	if c == nil {
		return nil
	}
	return &GcsModel{KeyId: optString(c.KeyId), Secret: optString(c.Secret)}
}

func s3FromAPI(c *client.S3Connection) *S3Model {
	if c == nil {
		return nil
	}
	return &S3Model{
		Provider:        optString(c.Provider),
		Chain:           optString(c.Chain),
		AccessKeyId:     optString(c.AccessKeyId),
		SecretAccessKey: optString(c.SecretAccessKey),
		Region:          optString(c.Region),
		Endpoint:        optString(c.Endpoint),
		SessionToken:    optString(c.SessionToken),
	}
}

func azureFromAPI(c *client.AzureConnection) *AzureModel {
	if c == nil {
		return nil
	}
	return &AzureModel{
		AuthType:     optString(c.AuthType),
		SasUrl:       optString(c.SasUrl),
		TenantId:     optString(c.TenantId),
		ClientId:     optString(c.ClientId),
		ClientSecret: optString(c.ClientSecret),
		AccountName:  optString(c.AccountName),
		FileUrl:      optString(c.FileUrl),
	}
}

// snapshotSecrets deep-copies the blocks that carry write-only credentials, so the
// values survive apiToModel rebuilding the model in place. Callers pass the same
// model as both source and destination (Create and Update both do), which makes a
// shallow copy alias the very pointers about to be overwritten.
//
// Only secret-bearing blocks are copied: everything else is refreshed from the API
// response and must not be carried over.
func snapshotSecrets(model, preserveSensitive *ConnectionResourceModel) *ConnectionResourceModel {
	src := preserveSensitive
	if src == nil {
		src = model
	}
	if src == nil {
		return nil
	}

	snap := &ConnectionResourceModel{}
	if p := src.Postgres; p != nil {
		snap.Postgres = &PostgresModel{Password: p.Password, ConnectionString: p.ConnectionString}
	}
	if b := src.Bigquery; b != nil {
		snap.Bigquery = &BigqueryModel{ServiceAccountKeyJson: b.ServiceAccountKeyJson}
	}
	if s := src.Snowflake; s != nil {
		snap.Snowflake = &SnowflakeModel{Password: s.Password, PrivateKey: s.PrivateKey, PrivateKeyPass: s.PrivateKeyPass}
	}
	if tr := src.Trino; tr != nil {
		snap.Trino = &TrinoModel{Password: tr.Password, PeakaKey: tr.PeakaKey}
	}
	if d := src.Databricks; d != nil {
		snap.Databricks = &DatabricksModel{Token: d.Token, OauthClientSecret: d.OauthClientSecret}
	}
	if m := src.Mysql; m != nil {
		snap.Mysql = &MysqlModel{Password: m.Password}
	}
	if md := src.Motherduck; md != nil {
		snap.Motherduck = &MotherduckModel{AccessToken: md.AccessToken}
	}
	if pb := src.Publisher; pb != nil {
		snap.Publisher = &PublisherModel{AccessToken: pb.AccessToken}
	}
	if dd := src.Duckdb; dd != nil {
		snap.Duckdb = &DuckdbModel{}
		for i := range dd.AttachedDatabases {
			ad := &dd.AttachedDatabases[i]
			snap.Duckdb.AttachedDatabases = append(snap.Duckdb.AttachedDatabases, AttachedDatabaseModel{
				Name:      ad.Name,
				Bigquery:  snapshotBigquerySecret(ad.Bigquery),
				Snowflake: snapshotSnowflakeSecrets(ad.Snowflake),
				Postgres:  snapshotPostgresSecrets(ad.Postgres),
				Gcs:       snapshotGcsSecret(ad.Gcs),
				S3:        snapshotS3Secrets(ad.S3),
				Azure:     snapshotAzureSecrets(ad.Azure),
			})
		}
	}
	if dl := src.Ducklake; dl != nil {
		snap.Ducklake = &DucklakeModel{}
		if s := dl.Storage; s != nil {
			snap.Ducklake.Storage = &DucklakeStorageModel{S3: snapshotS3Secrets(s.S3), Gcs: snapshotGcsSecret(s.Gcs)}
		}
		if c := dl.Catalog; c != nil {
			snap.Ducklake.Catalog = &DucklakeCatalogModel{Postgres: snapshotPostgresSecrets(c.Postgres)}
		}
	}
	if p := src.Proxy; p != nil && p.Ssh != nil {
		snap.Proxy = &ProxyModel{Ssh: &SshModel{PrivateKey: p.Ssh.PrivateKey, PrivateKeyPass: p.Ssh.PrivateKeyPass}}
	}
	return snap
}

func snapshotPostgresSecrets(m *PostgresModel) *PostgresModel {
	if m == nil {
		return nil
	}
	return &PostgresModel{Password: m.Password, ConnectionString: m.ConnectionString}
}

func snapshotBigquerySecret(m *BigqueryModel) *BigqueryModel {
	if m == nil {
		return nil
	}
	return &BigqueryModel{ServiceAccountKeyJson: m.ServiceAccountKeyJson}
}

func snapshotSnowflakeSecrets(m *SnowflakeModel) *SnowflakeModel {
	if m == nil {
		return nil
	}
	return &SnowflakeModel{Password: m.Password, PrivateKey: m.PrivateKey, PrivateKeyPass: m.PrivateKeyPass}
}

func snapshotGcsSecret(m *GcsModel) *GcsModel {
	if m == nil {
		return nil
	}
	return &GcsModel{Secret: m.Secret}
}

func snapshotS3Secrets(m *S3Model) *S3Model {
	if m == nil {
		return nil
	}
	return &S3Model{SecretAccessKey: m.SecretAccessKey, SessionToken: m.SessionToken}
}

func snapshotAzureSecrets(m *AzureModel) *AzureModel {
	if m == nil {
		return nil
	}
	return &AzureModel{ClientSecret: m.ClientSecret, SasUrl: m.SasUrl}
}

// clearEngineBlocks drops every engine block so the rebuild reflects only what the
// API returned. Without it, a connection whose engine changed out of band keeps its
// old block alongside the new one.
func clearEngineBlocks(model *ConnectionResourceModel) {
	model.Postgres = nil
	model.Bigquery = nil
	model.Snowflake = nil
	model.Trino = nil
	model.Databricks = nil
	model.Mysql = nil
	model.Duckdb = nil
	model.Motherduck = nil
	model.Ducklake = nil
	model.Publisher = nil
	model.Proxy = nil
}
