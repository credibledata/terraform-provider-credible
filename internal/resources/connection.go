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
	Project          types.String `tfsdk:"project"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	IncludeTables    types.List   `tfsdk:"include_tables"`
	ExcludeTables    types.List   `tfsdk:"exclude_tables"`
	ExcludeAllTables types.Bool   `tfsdk:"exclude_all_tables"`
	IndexingStatus   types.String `tfsdk:"indexing_status"`

	Postgres   *PostgresModel   `tfsdk:"postgres"`
	Bigquery   *BigqueryModel   `tfsdk:"bigquery"`
	Snowflake  *SnowflakeModel  `tfsdk:"snowflake"`
	Trino      *TrinoModel      `tfsdk:"trino"`
	Mysql      *MysqlModel      `tfsdk:"mysql"`
	Duckdb     *DuckdbModel     `tfsdk:"duckdb"`
	Motherduck *MotherduckModel `tfsdk:"motherduck"`
}

type PostgresModel struct {
	Host             types.String `tfsdk:"host"`
	Port             types.Int64  `tfsdk:"port"`
	DatabaseName     types.String `tfsdk:"database_name"`
	UserName         types.String `tfsdk:"user_name"`
	Password         types.String `tfsdk:"password"`
	ConnectionString types.String `tfsdk:"connection_string"`
}

type BigqueryModel struct {
	DefaultProjectId         types.String `tfsdk:"default_project_id"`
	BillingProjectId         types.String `tfsdk:"billing_project_id"`
	Location                 types.String `tfsdk:"location"`
	ServiceAccountKeyJson    types.String `tfsdk:"service_account_key_json"`
	MaximumBytesBilled       types.String `tfsdk:"maximum_bytes_billed"`
	QueryTimeoutMilliseconds types.String `tfsdk:"query_timeout_milliseconds"`
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
	Server  types.String `tfsdk:"server"`
	Port    types.Int64  `tfsdk:"port"`
	Catalog types.String `tfsdk:"catalog"`
	Schema  types.String `tfsdk:"schema"`
	User    types.String `tfsdk:"user"`
}

type MysqlModel struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	Database types.String `tfsdk:"database"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
}

type DuckdbModel struct {
	URL     types.String `tfsdk:"url"`
	MdToken types.String `tfsdk:"md_token"`
}

type MotherduckModel struct {
	URL     types.String `tfsdk:"url"`
	MdToken types.String `tfsdk:"md_token"`
}

func NewConnectionResource() resource.Resource {
	return &ConnectionResource{}
}

func (r *ConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection"
}

func (r *ConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a database connection within a Credible project.",
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
			"project": schema.StringAttribute{
				Description: "The project name.",
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
					stringvalidator.OneOf("postgres", "bigquery", "snowflake", "trino", "mysql", "duckdb", "motherduck"),
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
			"postgres": schema.SingleNestedBlock{
				Description: "PostgreSQL connection configuration.",
				Attributes: map[string]schema.Attribute{
					"host":              schema.StringAttribute{Optional: true, Description: "PostgreSQL server hostname."},
					"port":              schema.Int64Attribute{Optional: true, Description: "PostgreSQL server port."},
					"database_name":     schema.StringAttribute{Optional: true, Description: "Database name."},
					"user_name":         schema.StringAttribute{Optional: true, Description: "Username."},
					"password":          schema.StringAttribute{Optional: true, Sensitive: true, Description: "Password."},
					"connection_string": schema.StringAttribute{Optional: true, Sensitive: true, Description: "Full connection string (alternative to individual params)."},
				},
			},
			"bigquery": schema.SingleNestedBlock{
				Description: "BigQuery connection configuration.",
				Attributes: map[string]schema.Attribute{
					"default_project_id":        schema.StringAttribute{Optional: true, Description: "Default BigQuery project ID."},
					"billing_project_id":        schema.StringAttribute{Optional: true, Description: "Billing project ID."},
					"location":                  schema.StringAttribute{Optional: true, Description: "Dataset location."},
					"service_account_key_json":  schema.StringAttribute{Optional: true, Sensitive: true, Description: "Service account key JSON."},
					"maximum_bytes_billed":      schema.StringAttribute{Optional: true, Description: "Maximum bytes billed."},
					"query_timeout_milliseconds": schema.StringAttribute{Optional: true, Description: "Query timeout in milliseconds."},
				},
			},
			"snowflake": schema.SingleNestedBlock{
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
			},
			"trino": schema.SingleNestedBlock{
				Description: "Trino connection configuration.",
				Attributes: map[string]schema.Attribute{
					"server":  schema.StringAttribute{Optional: true, Description: "Trino server hostname."},
					"port":    schema.Int64Attribute{Optional: true, Description: "Trino server port."},
					"catalog": schema.StringAttribute{Optional: true, Description: "Catalog name."},
					"schema":  schema.StringAttribute{Optional: true, Description: "Schema name."},
					"user":    schema.StringAttribute{Optional: true, Description: "Username."},
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
				Description: "DuckDB connection configuration.",
				Attributes: map[string]schema.Attribute{
					"url":      schema.StringAttribute{Optional: true, Description: "DuckDB URL."},
					"md_token": schema.StringAttribute{Optional: true, Sensitive: true, Description: "MotherDuck token."},
				},
			},
			"motherduck": schema.SingleNestedBlock{
				Description: "MotherDuck connection configuration.",
				Attributes: map[string]schema.Attribute{
					"url":      schema.StringAttribute{Optional: true, Description: "MotherDuck URL."},
					"md_token": schema.StringAttribute{Optional: true, Sensitive: true, Description: "MotherDuck token."},
				},
			},
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
	if model.Postgres != nil {
		conn.PostgresConnection = &client.PostgresConnection{}
		if !model.Postgres.Host.IsNull() {
			conn.PostgresConnection.Host = model.Postgres.Host.ValueString()
		}
		if !model.Postgres.Port.IsNull() {
			p := int(model.Postgres.Port.ValueInt64())
			conn.PostgresConnection.Port = &p
		}
		if !model.Postgres.DatabaseName.IsNull() {
			conn.PostgresConnection.DatabaseName = model.Postgres.DatabaseName.ValueString()
		}
		if !model.Postgres.UserName.IsNull() {
			conn.PostgresConnection.UserName = model.Postgres.UserName.ValueString()
		}
		if !model.Postgres.Password.IsNull() {
			conn.PostgresConnection.Password = model.Postgres.Password.ValueString()
		}
		if !model.Postgres.ConnectionString.IsNull() {
			conn.PostgresConnection.ConnectionString = model.Postgres.ConnectionString.ValueString()
		}
	}

	if model.Bigquery != nil {
		conn.BigqueryConnection = &client.BigqueryConnection{}
		if !model.Bigquery.DefaultProjectId.IsNull() {
			conn.BigqueryConnection.DefaultProjectId = model.Bigquery.DefaultProjectId.ValueString()
		}
		if !model.Bigquery.BillingProjectId.IsNull() {
			conn.BigqueryConnection.BillingProjectId = model.Bigquery.BillingProjectId.ValueString()
		}
		if !model.Bigquery.Location.IsNull() {
			conn.BigqueryConnection.Location = model.Bigquery.Location.ValueString()
		}
		if !model.Bigquery.ServiceAccountKeyJson.IsNull() {
			conn.BigqueryConnection.ServiceAccountKeyJson = model.Bigquery.ServiceAccountKeyJson.ValueString()
		}
		if !model.Bigquery.MaximumBytesBilled.IsNull() {
			conn.BigqueryConnection.MaximumBytesBilled = model.Bigquery.MaximumBytesBilled.ValueString()
		}
		if !model.Bigquery.QueryTimeoutMilliseconds.IsNull() {
			conn.BigqueryConnection.QueryTimeoutMilliseconds = model.Bigquery.QueryTimeoutMilliseconds.ValueString()
		}
	}

	if model.Snowflake != nil {
		conn.SnowflakeConnection = &client.SnowflakeConnection{}
		if !model.Snowflake.Account.IsNull() {
			conn.SnowflakeConnection.Account = model.Snowflake.Account.ValueString()
		}
		if !model.Snowflake.Username.IsNull() {
			conn.SnowflakeConnection.Username = model.Snowflake.Username.ValueString()
		}
		if !model.Snowflake.Password.IsNull() {
			conn.SnowflakeConnection.Password = model.Snowflake.Password.ValueString()
		}
		if !model.Snowflake.PrivateKey.IsNull() {
			conn.SnowflakeConnection.PrivateKey = model.Snowflake.PrivateKey.ValueString()
		}
		if !model.Snowflake.PrivateKeyPass.IsNull() {
			conn.SnowflakeConnection.PrivateKeyPass = model.Snowflake.PrivateKeyPass.ValueString()
		}
		if !model.Snowflake.Warehouse.IsNull() {
			conn.SnowflakeConnection.Warehouse = model.Snowflake.Warehouse.ValueString()
		}
		if !model.Snowflake.Database.IsNull() {
			conn.SnowflakeConnection.Database = model.Snowflake.Database.ValueString()
		}
		if !model.Snowflake.Schema.IsNull() {
			conn.SnowflakeConnection.Schema = model.Snowflake.Schema.ValueString()
		}
		if !model.Snowflake.Role.IsNull() {
			conn.SnowflakeConnection.Role = model.Snowflake.Role.ValueString()
		}
		if !model.Snowflake.ResponseTimeoutMilliseconds.IsNull() {
			t := int(model.Snowflake.ResponseTimeoutMilliseconds.ValueInt64())
			conn.SnowflakeConnection.ResponseTimeoutMilliseconds = &t
		}
	}

	if model.Trino != nil {
		conn.TrinoConnection = &client.TrinoConnection{}
		if !model.Trino.Server.IsNull() {
			conn.TrinoConnection.Server = model.Trino.Server.ValueString()
		}
		if !model.Trino.Port.IsNull() {
			p := int(model.Trino.Port.ValueInt64())
			conn.TrinoConnection.Port = &p
		}
		if !model.Trino.Catalog.IsNull() {
			conn.TrinoConnection.Catalog = model.Trino.Catalog.ValueString()
		}
		if !model.Trino.Schema.IsNull() {
			conn.TrinoConnection.Schema = model.Trino.Schema.ValueString()
		}
		if !model.Trino.User.IsNull() {
			conn.TrinoConnection.User = model.Trino.User.ValueString()
		}
	}

	if model.Mysql != nil {
		conn.MysqlConnection = &client.MysqlConnection{}
		if !model.Mysql.Host.IsNull() {
			conn.MysqlConnection.Host = model.Mysql.Host.ValueString()
		}
		if !model.Mysql.Port.IsNull() {
			p := int(model.Mysql.Port.ValueInt64())
			conn.MysqlConnection.Port = &p
		}
		if !model.Mysql.Database.IsNull() {
			conn.MysqlConnection.Database = model.Mysql.Database.ValueString()
		}
		if !model.Mysql.User.IsNull() {
			conn.MysqlConnection.User = model.Mysql.User.ValueString()
		}
		if !model.Mysql.Password.IsNull() {
			conn.MysqlConnection.Password = model.Mysql.Password.ValueString()
		}
	}

	if model.Duckdb != nil {
		conn.DuckdbConnection = &client.DuckdbConnection{}
		if !model.Duckdb.URL.IsNull() {
			conn.DuckdbConnection.URL = model.Duckdb.URL.ValueString()
		}
		if !model.Duckdb.MdToken.IsNull() {
			conn.DuckdbConnection.MdToken = model.Duckdb.MdToken.ValueString()
		}
	}

	if model.Motherduck != nil {
		conn.MotherduckConnection = &client.MotherduckConnection{}
		if !model.Motherduck.URL.IsNull() {
			conn.MotherduckConnection.URL = model.Motherduck.URL.ValueString()
		}
		if !model.Motherduck.MdToken.IsNull() {
			conn.MotherduckConnection.MdToken = model.Motherduck.MdToken.ValueString()
		}
	}

	return conn
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
	if preserveSensitive != nil && preserveSensitive.Duckdb != nil && model.Duckdb != nil {
		if !preserveSensitive.Duckdb.MdToken.IsNull() {
			model.Duckdb.MdToken = preserveSensitive.Duckdb.MdToken
		}
	}
	if preserveSensitive != nil && preserveSensitive.Motherduck != nil && model.Motherduck != nil {
		if !preserveSensitive.Motherduck.MdToken.IsNull() {
			model.Motherduck.MdToken = preserveSensitive.Motherduck.MdToken
		}
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

	tflog.Debug(ctx, "Creating connection", map[string]interface{}{"org": org, "project": plan.Project.ValueString(), "name": conn.Name})

	_, err := r.client.CreateConnection(org, plan.Project.ValueString(), conn)
	if err != nil {
		resp.Diagnostics.AddError("Error creating connection", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)

	// Read back to get computed fields
	result, err := r.client.GetConnection(org, plan.Project.ValueString(), plan.Name.ValueString())
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
	result, err := r.client.GetConnection(org, state.Project.ValueString(), state.Name.ValueString())
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

	_, err := r.client.UpdateConnection(org, plan.Project.ValueString(), plan.Name.ValueString(), conn)
	if err != nil {
		resp.Diagnostics.AddError("Error updating connection", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)

	result, err := r.client.GetConnection(org, plan.Project.ValueString(), plan.Name.ValueString())
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
	err := r.client.DeleteConnection(org, state.Project.ValueString(), state.Name.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting connection", err.Error())
	}
}

func (r *ConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be in the format: organization/project/connection")
		return
	}

	org, project, name := parts[0], parts[1], parts[2]
	result, err := r.client.GetConnection(org, project, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing connection", err.Error())
		return
	}

	state := ConnectionResourceModel{
		Organization:   types.StringValue(org),
		Project:        types.StringValue(project),
		Name:           types.StringValue(result.Name),
		Type:           types.StringValue(result.Type),
		IndexingStatus: types.StringValue(result.IndexingStatus),
	}
	if result.ExcludeAllTables != nil {
		state.ExcludeAllTables = types.BoolValue(*result.ExcludeAllTables)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
