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

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

type ProjectResource struct {
	client *client.Client
}

type ProjectResourceModel struct {
	Organization       types.String `tfsdk:"organization"`
	Name               types.String `tfsdk:"name"`
	Readme             types.String `tfsdk:"readme"`
	ReplicationCount   types.Int64  `tfsdk:"replication_count"`
	DeletionProtection types.Bool   `tfsdk:"deletion_protection"`
	ForceCascade       types.Bool   `tfsdk:"force_cascade"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

func (r *ProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Credible project within an organization.",
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
				Description: "The unique name of the project.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"readme": schema.StringAttribute{
				Description: "Markdown-formatted project description.",
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
				Description: "Whether deletion protection is enabled. Must be set to false before the project can be destroyed. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"force_cascade": schema.BoolAttribute{
				Description: "If true, allow deleting the project even if it contains packages or connections. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"created_at": schema.StringAttribute{
				Description: "When the project was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the project was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *ProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectResource) getOrg(model *ProjectResourceModel) string {
	if !model.Organization.IsNull() && !model.Organization.IsUnknown() {
		return model.Organization.ValueString()
	}
	return r.client.Organization
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	if org == "" {
		resp.Diagnostics.AddError("Missing organization", "Organization must be set either on the resource or provider.")
		return
	}

	project := &client.Project{
		Name: plan.Name.ValueString(),
	}
	if !plan.Readme.IsNull() {
		project.Readme = plan.Readme.ValueString()
	}
	if !plan.ReplicationCount.IsNull() && !plan.ReplicationCount.IsUnknown() {
		rc := int(plan.ReplicationCount.ValueInt64())
		project.ReplicationCount = &rc
	}

	tflog.Debug(ctx, "Creating project", map[string]interface{}{"org": org, "name": project.Name})

	result, err := r.client.CreateProject(org, project)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
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

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	result, err := r.client.GetProject(org, state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project", err.Error())
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

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	project := &client.Project{}
	if !plan.Readme.IsNull() {
		project.Readme = plan.Readme.ValueString()
	}
	if !plan.ReplicationCount.IsNull() && !plan.ReplicationCount.IsUnknown() {
		rc := int(plan.ReplicationCount.ValueInt64())
		project.ReplicationCount = &rc
	}

	result, err := r.client.UpdateProject(org, plan.Name.ValueString(), project)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
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

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.DeletionProtection.ValueBool() {
		resp.Diagnostics.AddError(
			"Project is protected",
			fmt.Sprintf("Project %q has deletion_protection = true. Set it to false before destroying.", state.Name.ValueString()),
		)
		return
	}

	org := r.getOrg(&state)
	name := state.Name.ValueString()

	if !state.ForceCascade.ValueBool() {
		packages, err := r.client.ListPackages(org, name)
		if err != nil {
			resp.Diagnostics.AddError("Error checking project contents", err.Error())
			return
		}
		connections, err := r.client.ListConnections(org, name)
		if err != nil {
			resp.Diagnostics.AddError("Error checking project contents", err.Error())
			return
		}
		if len(packages) > 0 || len(connections) > 0 {
			resp.Diagnostics.AddError(
				"Project is not empty",
				fmt.Sprintf("Project %q contains %d package(s) and %d connection(s). Set force_cascade = true to allow deletion, or remove them first.", name, len(packages), len(connections)),
			)
			return
		}
	}

	err := r.client.DeleteProject(org, name)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting project", err.Error())
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be in the format: organization/project")
		return
	}

	org, name := parts[0], parts[1]
	result, err := r.client.GetProject(org, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing project", err.Error())
		return
	}

	state := ProjectResourceModel{
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
