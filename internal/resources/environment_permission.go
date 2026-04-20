package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/credibledata/terraform-provider-credible/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EnvironmentPermissionResource{}
var _ resource.ResourceWithImportState = &EnvironmentPermissionResource{}

type EnvironmentPermissionResource struct {
	client *client.Client
}

type EnvironmentPermissionResourceModel struct {
	Organization types.String `tfsdk:"organization"`
	Environment  types.String `tfsdk:"environment"`
	UserGroupID  types.String `tfsdk:"user_group_id"`
	Permission   types.String `tfsdk:"permission"`
}

func NewEnvironmentPermissionResource() resource.Resource {
	return &EnvironmentPermissionResource{}
}

func (r *EnvironmentPermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_permission"
}

func (r *EnvironmentPermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a permission assignment for a user or group within a Credible environment.",
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

func (r *EnvironmentPermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnvironmentPermissionResource) getOrg(model *EnvironmentPermissionResourceModel) string {
	if !model.Organization.IsNull() && !model.Organization.IsUnknown() {
		return model.Organization.ValueString()
	}
	return r.client.Organization
}

func (r *EnvironmentPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentPermissionResourceModel
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

	result, err := r.client.CreateEnvironmentPermission(org, plan.Environment.ValueString(), perm)
	if err != nil {
		resp.Diagnostics.AddError("Error creating environment permission", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.UserGroupID = types.StringValue(result.UserGroupID)
	plan.Permission = types.StringValue(result.Permission)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentPermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	result, err := r.client.GetEnvironmentPermission(org, state.Environment.ValueString(), state.UserGroupID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading environment permission", err.Error())
		return
	}

	state.Organization = types.StringValue(org)
	state.UserGroupID = types.StringValue(result.UserGroupID)
	state.Permission = types.StringValue(result.Permission)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EnvironmentPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentPermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	perm := &client.Permission{
		Permission: plan.Permission.ValueString(),
	}

	result, err := r.client.UpdateEnvironmentPermission(org, plan.Environment.ValueString(), plan.UserGroupID.ValueString(), perm)
	if err != nil {
		resp.Diagnostics.AddError("Error updating environment permission", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.UserGroupID = types.StringValue(result.UserGroupID)
	plan.Permission = types.StringValue(result.Permission)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentPermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	err := r.client.DeleteEnvironmentPermission(org, state.Environment.ValueString(), state.UserGroupID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting environment permission", err.Error())
	}
}

func (r *EnvironmentPermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be in the format: organization/environment/user_group_id")
		return
	}

	org, environment, userGroupID := parts[0], parts[1], parts[2]
	result, err := r.client.GetEnvironmentPermission(org, environment, userGroupID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing environment permission", err.Error())
		return
	}

	state := EnvironmentPermissionResourceModel{
		Organization: types.StringValue(org),
		Environment:  types.StringValue(environment),
		UserGroupID:  types.StringValue(result.UserGroupID),
		Permission:   types.StringValue(result.Permission),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
