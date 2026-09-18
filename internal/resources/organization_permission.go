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

// toAPI builds the request body for a create or an update.
//
// The subject rides along with the role on both. The server addresses the grant
// by path but validates the body, so a body naming no subject is refused -- and
// a role change is an in-place update, so this is the only request a managed
// permission sends after its create.
func (m *OrganizationPermissionResourceModel) toAPI() *client.Permission {
	return &client.Permission{
		UserGroupID: m.UserGroupID.ValueString(),
		Permission:  m.Permission.ValueString(),
	}
}

// applyResult writes the API's view of the grant onto the model.
//
// The subject is the caller's, never the response's: it is this resource's
// identity, the API is asked for it by path, and a response that omits it would
// empty the id in state. An emptied id makes the next refresh request
// `.../permissions/` with no id, which 404s, drops the resource and recreates it
// into a 409 against the grant that is still there.
func (m *OrganizationPermissionResourceModel) applyResult(org, userGroupID string, result *client.Permission) {
	m.Organization = types.StringValue(org)
	m.UserGroupID = types.StringValue(userGroupID)
	m.Permission = types.StringValue(result.Permission)
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

	result, err := r.client.CreateOrgPermission(org, plan.toAPI())
	if err != nil {
		resp.Diagnostics.AddError("Error creating organization permission", err.Error())
		return
	}

	plan.applyResult(org, plan.UserGroupID.ValueString(), result)

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

	state.applyResult(org, state.UserGroupID.ValueString(), result)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *OrganizationPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OrganizationPermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)

	result, err := r.client.UpdateOrgPermission(org, plan.UserGroupID.ValueString(), plan.toAPI())
	if err != nil {
		resp.Diagnostics.AddError("Error updating organization permission", err.Error())
		return
	}

	plan.applyResult(org, plan.UserGroupID.ValueString(), result)

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

	var state OrganizationPermissionResourceModel
	state.applyResult(org, userGroupID, result)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
