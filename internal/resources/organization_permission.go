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

var _ resource.Resource = &OrganizationPermissionResource{}
var _ resource.ResourceWithImportState = &OrganizationPermissionResource{}

type OrganizationPermissionResource struct {
	client *client.Client
}

type OrganizationPermissionResourceModel struct {
	Organization types.String `tfsdk:"organization"`
	UserGroupID  types.String `tfsdk:"user_group_id"`
	Permission   types.String `tfsdk:"permission"`
}

func NewOrganizationPermissionResource() resource.Resource {
	return &OrganizationPermissionResource{}
}

func (r *OrganizationPermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_permission"
}

func (r *OrganizationPermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a permission assignment for a user or group within a Credible organization.",
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
			"user_group_id": schema.StringAttribute{
				Description: "The user or group identifier. Format: 'user:{email}' or 'group:{groupName}'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permission": schema.StringAttribute{
				Description: "The permission level: admin, modeler, or member.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("admin", "modeler", "member"),
				},
			},
		},
	}
}

func (r *OrganizationPermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationPermissionResource) getOrg(model *OrganizationPermissionResourceModel) string {
	if !model.Organization.IsNull() && !model.Organization.IsUnknown() {
		return model.Organization.ValueString()
	}
	return r.client.Organization
}

func (r *OrganizationPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OrganizationPermissionResourceModel
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

	result, err := r.client.CreateOrgPermission(org, perm)
	if err != nil {
		resp.Diagnostics.AddError("Error creating organization permission", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.UserGroupID = types.StringValue(result.UserGroupID)
	plan.Permission = types.StringValue(result.Permission)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OrganizationPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OrganizationPermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	result, err := r.client.GetOrgPermission(org, state.UserGroupID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading organization permission", err.Error())
		return
	}

	state.Organization = types.StringValue(org)
	state.UserGroupID = types.StringValue(result.UserGroupID)
	state.Permission = types.StringValue(result.Permission)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *OrganizationPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OrganizationPermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	perm := &client.Permission{
		Permission: plan.Permission.ValueString(),
	}

	result, err := r.client.UpdateOrgPermission(org, plan.UserGroupID.ValueString(), perm)
	if err != nil {
		resp.Diagnostics.AddError("Error updating organization permission", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.UserGroupID = types.StringValue(result.UserGroupID)
	plan.Permission = types.StringValue(result.Permission)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OrganizationPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OrganizationPermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	err := r.client.DeleteOrgPermission(org, state.UserGroupID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting organization permission", err.Error())
	}
}

func (r *OrganizationPermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Format: org/user:email@example.com or org/group:groupname
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be in the format: organization/user_group_id (e.g., my-org/user:alice@example.com)")
		return
	}

	org, userGroupID := parts[0], parts[1]
	result, err := r.client.GetOrgPermission(org, userGroupID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing organization permission", err.Error())
		return
	}

	state := OrganizationPermissionResourceModel{
		Organization: types.StringValue(org),
		UserGroupID:  types.StringValue(result.UserGroupID),
		Permission:   types.StringValue(result.Permission),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
