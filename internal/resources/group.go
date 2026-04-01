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
)

var _ resource.Resource = &GroupResource{}
var _ resource.ResourceWithImportState = &GroupResource{}

type GroupResource struct {
	client *client.Client
}

type GroupResourceModel struct {
	Organization types.String `tfsdk:"organization"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

func (r *GroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a group within a Credible organization.",
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
				Description: "The group name. Must be unique within the organization.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A human-readable description of the group.",
				Optional:    true,
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the group was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the group was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *GroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupResource) getOrg(model *GroupResourceModel) string {
	if !model.Organization.IsNull() && !model.Organization.IsUnknown() {
		return model.Organization.ValueString()
	}
	return r.client.Organization
}

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	if org == "" {
		resp.Diagnostics.AddError("Missing organization", "Organization must be set either on the resource or provider.")
		return
	}

	group := &client.Group{
		GroupName:   plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	result, err := r.client.CreateGroup(org, group)
	if err != nil {
		resp.Diagnostics.AddError("Error creating group", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.Name = types.StringValue(result.GroupName)
	plan.Description = types.StringValue(result.Description)
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	result, err := r.client.GetGroup(org, state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading group", err.Error())
		return
	}

	state.Organization = types.StringValue(org)
	state.Name = types.StringValue(result.GroupName)
	state.Description = types.StringValue(result.Description)
	state.CreatedAt = types.StringValue(result.CreatedAt)
	state.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	group := &client.Group{
		Description: plan.Description.ValueString(),
	}

	result, err := r.client.UpdateGroup(org, plan.Name.ValueString(), group)
	if err != nil {
		resp.Diagnostics.AddError("Error updating group", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.Name = types.StringValue(result.GroupName)
	plan.Description = types.StringValue(result.Description)
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	err := r.client.DeleteGroup(org, state.Name.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting group", err.Error())
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Format: org/groupName
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be in the format: organization/group_name")
		return
	}

	org, groupName := parts[0], parts[1]
	result, err := r.client.GetGroup(org, groupName)
	if err != nil {
		resp.Diagnostics.AddError("Error importing group", err.Error())
		return
	}

	state := GroupResourceModel{
		Organization: types.StringValue(org),
		Name:         types.StringValue(result.GroupName),
		Description:  types.StringValue(result.Description),
		CreatedAt:    types.StringValue(result.CreatedAt),
		UpdatedAt:    types.StringValue(result.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
