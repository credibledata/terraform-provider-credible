package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/credibledata/terraform-provider-credible/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

var _ resource.Resource = &ProjectPermissionResource{}
var _ resource.ResourceWithImportState = &ProjectPermissionResource{}

type ProjectPermissionResource struct {
	client *client.Client
}

type ProjectPermissionResourceModel struct {
	Organization types.String `tfsdk:"organization"`
	Project      types.String `tfsdk:"project"`
	UserGroupID  types.String `tfsdk:"user_group_id"`
	Permission   types.String `tfsdk:"permission"`
}

func NewProjectPermissionResource() resource.Resource {
	return &ProjectPermissionResource{}
}

func (r *ProjectPermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_permission"
}

func (r *ProjectPermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a permission assignment for a user or group within a Credible project.",
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
			"user_group_id": schema.StringAttribute{
				Description: "The user or group identifier. Format: 'user:{email}' or 'group:{groupName}'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permission": schema.StringAttribute{
				Description: "The permission level: admin, modeler, or viewer.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("admin", "modeler", "viewer"),
				},
			},
		},
	}
}

func (r *ProjectPermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectPermissionResource) getOrg(model *ProjectPermissionResourceModel) string {
	if !model.Organization.IsNull() && !model.Organization.IsUnknown() {
		return model.Organization.ValueString()
	}
	return r.client.Organization
}

func (r *ProjectPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectPermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	if org == "" {
		resp.Diagnostics.AddError("Missing organization", "Organization must be set either on the resource or provider.")
		return
	}

	perm := &client.Permission{
		UserGroupID: plan.UserGroupID.ValueString(),
		Permission:  plan.Permission.ValueString(),
	}

	result, err := r.client.CreateProjectPermission(org, plan.Project.ValueString(), perm)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project permission", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.UserGroupID = types.StringValue(result.UserGroupID)
	plan.Permission = types.StringValue(result.Permission)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectPermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	result, err := r.client.GetProjectPermission(org, state.Project.ValueString(), state.UserGroupID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project permission", err.Error())
		return
	}

	state.Organization = types.StringValue(org)
	state.UserGroupID = types.StringValue(result.UserGroupID)
	state.Permission = types.StringValue(result.Permission)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProjectPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectPermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	perm := &client.Permission{
		Permission: plan.Permission.ValueString(),
	}

	result, err := r.client.UpdateProjectPermission(org, plan.Project.ValueString(), plan.UserGroupID.ValueString(), perm)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project permission", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.UserGroupID = types.StringValue(result.UserGroupID)
	plan.Permission = types.StringValue(result.Permission)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectPermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	err := r.client.DeleteProjectPermission(org, state.Project.ValueString(), state.UserGroupID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting project permission", err.Error())
	}
}

func (r *ProjectPermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Format: org/project/user:email@example.com
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be in the format: organization/project/user_group_id")
		return
	}

	org, project, userGroupID := parts[0], parts[1], parts[2]
	result, err := r.client.GetProjectPermission(org, project, userGroupID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing project permission", err.Error())
		return
	}

	state := ProjectPermissionResourceModel{
		Organization: types.StringValue(org),
		Project:      types.StringValue(project),
		UserGroupID:  types.StringValue(result.UserGroupID),
		Permission:   types.StringValue(result.Permission),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
