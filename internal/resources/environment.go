package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/credibledata/terraform-provider-credible/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &EnvironmentResource{}
var _ resource.ResourceWithImportState = &EnvironmentResource{}

type EnvironmentResource struct {
	client *client.Client
}

type EnvironmentResourceModel struct {
	Organization       types.String `tfsdk:"organization"`
	Name               types.String `tfsdk:"name"`
	Readme             types.String `tfsdk:"readme"`
	ReplicationCount   types.Int64  `tfsdk:"replication_count"`
	DeletionProtection types.Bool   `tfsdk:"deletion_protection"`
	ForceCascade       types.Bool   `tfsdk:"force_cascade"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

func NewEnvironmentResource() resource.Resource {
	return &EnvironmentResource{}
}

func (r *EnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *EnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Credible environment within an organization.",
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
			"name": schema.StringAttribute{
				Description: "The unique name of the environment.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"readme": schema.StringAttribute{
				Description: "Markdown-formatted environment description.",
				Optional:    true,
			},
			"replication_count": schema.Int64Attribute{
				Description: "Number of replicas for high availability (1-10).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"deletion_protection": schema.BoolAttribute{
				Description: "Whether deletion protection is enabled. Must be set to false before the environment can be destroyed. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"force_cascade": schema.BoolAttribute{
				Description: "If true, allow deleting the environment even if it contains packages or connections. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"created_at": schema.StringAttribute{
				Description: "When the environment was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the environment was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *EnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnvironmentResource) getOrg(model *EnvironmentResourceModel) string {
	if !model.Organization.IsNull() && !model.Organization.IsUnknown() {
		return model.Organization.ValueString()
	}
	return r.client.Organization
}

func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	if org == "" {
		resp.Diagnostics.AddError("Missing organization", "Organization must be set either on the resource or provider.")
		return
	}

	env := &client.Environment{
		Name: plan.Name.ValueString(),
	}
	if !plan.Readme.IsNull() {
		env.Readme = plan.Readme.ValueString()
	}
	if !plan.ReplicationCount.IsNull() && !plan.ReplicationCount.IsUnknown() {
		rc := int(plan.ReplicationCount.ValueInt64())
		env.ReplicationCount = &rc
	}

	tflog.Debug(ctx, "Creating environment", map[string]interface{}{"org": org, "name": env.Name})

	result, err := r.client.CreateEnvironment(org, env)
	if err != nil {
		resp.Diagnostics.AddError("Error creating environment", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.Name = types.StringValue(result.Name)
	if result.Readme != "" {
		plan.Readme = types.StringValue(result.Readme)
	}
	if result.ReplicationCount != nil {
		plan.ReplicationCount = types.Int64Value(int64(*result.ReplicationCount))
	}
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	result, err := r.client.GetEnvironment(org, state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading environment", err.Error())
		return
	}

	state.Organization = types.StringValue(org)
	state.Name = types.StringValue(result.Name)
	state.Readme = types.StringValue(result.Readme)
	if result.ReplicationCount != nil {
		state.ReplicationCount = types.Int64Value(int64(*result.ReplicationCount))
	}
	state.CreatedAt = types.StringValue(result.CreatedAt)
	state.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	env := &client.Environment{}
	if !plan.Readme.IsNull() {
		env.Readme = plan.Readme.ValueString()
	}
	if !plan.ReplicationCount.IsNull() && !plan.ReplicationCount.IsUnknown() {
		rc := int(plan.ReplicationCount.ValueInt64())
		env.ReplicationCount = &rc
	}

	result, err := r.client.UpdateEnvironment(org, plan.Name.ValueString(), env)
	if err != nil {
		resp.Diagnostics.AddError("Error updating environment", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.Name = types.StringValue(result.Name)
	plan.Readme = types.StringValue(result.Readme)
	if result.ReplicationCount != nil {
		plan.ReplicationCount = types.Int64Value(int64(*result.ReplicationCount))
	}
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.DeletionProtection.ValueBool() {
		resp.Diagnostics.AddError(
			"Environment is protected",
			fmt.Sprintf("Environment %q has deletion_protection = true. Set it to false before destroying.", state.Name.ValueString()),
		)
		return
	}

	org := r.getOrg(&state)
	name := state.Name.ValueString()

	if !state.ForceCascade.ValueBool() {
		packages, err := r.client.ListPackages(org, name)
		if err != nil {
			resp.Diagnostics.AddError("Error checking environment contents", err.Error())
			return
		}
		connections, err := r.client.ListConnections(org, name)
		if err != nil {
			resp.Diagnostics.AddError("Error checking environment contents", err.Error())
			return
		}
		if len(packages) > 0 || len(connections) > 0 {
			resp.Diagnostics.AddError(
				"Environment is not empty",
				fmt.Sprintf("Environment %q contains %d package(s) and %d connection(s). Set force_cascade = true to allow deletion, or remove them first.", name, len(packages), len(connections)),
			)
			return
		}
	}

	err := r.client.DeleteEnvironment(org, name)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting environment", err.Error())
	}
}

func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be in the format: organization/environment")
		return
	}

	org, name := parts[0], parts[1]
	result, err := r.client.GetEnvironment(org, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing environment", err.Error())
		return
	}

	state := EnvironmentResourceModel{
		Organization: types.StringValue(org),
		Name:         types.StringValue(result.Name),
		Readme:       types.StringValue(result.Readme),
		CreatedAt:    types.StringValue(result.CreatedAt),
		UpdatedAt:    types.StringValue(result.UpdatedAt),
	}
	if result.ReplicationCount != nil {
		state.ReplicationCount = types.Int64Value(int64(*result.ReplicationCount))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
