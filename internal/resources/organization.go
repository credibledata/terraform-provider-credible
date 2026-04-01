package resources

import (
	"context"
	"fmt"

	"github.com/credibledata/terraform-provider-credible/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &OrganizationResource{}
var _ resource.ResourceWithImportState = &OrganizationResource{}

type OrganizationResource struct {
	client *client.Client
}

type OrganizationResourceModel struct {
	Name               types.String `tfsdk:"name"`
	DisplayName        types.String `tfsdk:"display_name"`
	DeletionProtection types.Bool   `tfsdk:"deletion_protection"`
	ForceCascade       types.Bool   `tfsdk:"force_cascade"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

func NewOrganizationResource() resource.Resource {
	return &OrganizationResource{}
}

func (r *OrganizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *OrganizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Credible organization.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The unique name of the organization.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the organization.",
				Optional:    true,
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the organization was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the organization was last updated.",
				Computed:    true,
			},
			"deletion_protection": schema.BoolAttribute{
				Description: "Whether deletion protection is enabled. Must be set to false before the organization can be destroyed. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"force_cascade": schema.BoolAttribute{
				Description: "If true, allow deleting the organization even if it contains projects. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *OrganizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OrganizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := &client.Organization{
		Name: plan.Name.ValueString(),
	}
	if !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() {
		org.DisplayName = plan.DisplayName.ValueString()
	}

	tflog.Debug(ctx, "Creating organization", map[string]interface{}{"name": org.Name})

	result, err := r.client.CreateOrganization(org)
	if err != nil {
		resp.Diagnostics.AddError("Error creating organization", err.Error())
		return
	}

	plan.Name = types.StringValue(result.Name)
	plan.DisplayName = types.StringValue(result.DisplayName)
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OrganizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OrganizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetOrganization(state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading organization", err.Error())
		return
	}

	state.Name = types.StringValue(result.Name)
	state.DisplayName = types.StringValue(result.DisplayName)
	state.CreatedAt = types.StringValue(result.CreatedAt)
	state.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *OrganizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OrganizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := &client.Organization{}
	if !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() {
		org.DisplayName = plan.DisplayName.ValueString()
	}

	result, err := r.client.UpdateOrganization(plan.Name.ValueString(), org)
	if err != nil {
		resp.Diagnostics.AddError("Error updating organization", err.Error())
		return
	}

	plan.Name = types.StringValue(result.Name)
	plan.DisplayName = types.StringValue(result.DisplayName)
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *OrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OrganizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.DeletionProtection.ValueBool() {
		resp.Diagnostics.AddError(
			"Organization is protected",
			fmt.Sprintf("Organization %q has deletion_protection = true. Set it to false before destroying.", state.Name.ValueString()),
		)
		return
	}

	if !state.ForceCascade.ValueBool() {
		projects, err := r.client.ListProjects(state.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error checking organization contents", err.Error())
			return
		}
		if len(projects) > 0 {
			resp.Diagnostics.AddError(
				"Organization is not empty",
				fmt.Sprintf("Organization %q contains %d project(s). Set force_cascade = true to allow deletion, or remove the projects first.", state.Name.ValueString(), len(projects)),
			)
			return
		}
	}

	// The Credible API does not have a DELETE endpoint for organizations.
	// Remove from state only.
	tflog.Warn(ctx, "Organization delete is not supported by the API. Removing from Terraform state only.")
}

func (r *OrganizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	result, err := r.client.GetOrganization(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing organization", err.Error())
		return
	}

	state := OrganizationResourceModel{
		Name:        types.StringValue(result.Name),
		DisplayName: types.StringValue(result.DisplayName),
		CreatedAt:   types.StringValue(result.CreatedAt),
		UpdatedAt:   types.StringValue(result.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
