package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/credibledata/terraform-provider-credible/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &PackageResource{}
var _ resource.ResourceWithImportState = &PackageResource{}

type PackageResource struct {
	client *client.Client
}

type PackageResourceModel struct {
	Organization       types.String `tfsdk:"organization"`
	Project            types.String `tfsdk:"project"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	DeletionProtection types.Bool   `tfsdk:"deletion_protection"`
	LatestVersion      types.String `tfsdk:"latest_version"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

func NewPackageResource() resource.Resource {
	return &PackageResource{}
}

func (r *PackageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package"
}

func (r *PackageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Credible package within a project.",
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
				Description: "The unique name of the package.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the package.",
				Optional:    true,
			},
			"deletion_protection": schema.BoolAttribute{
				Description: "Whether deletion protection is enabled. Must be set to false before the package can be destroyed. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"latest_version": schema.StringAttribute{
				Description: "The latest published version.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the package was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the package was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *PackageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PackageResource) getOrg(model *PackageResourceModel) string {
	if !model.Organization.IsNull() && !model.Organization.IsUnknown() {
		return model.Organization.ValueString()
	}
	return r.client.Organization
}

func (r *PackageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PackageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	if org == "" {
		resp.Diagnostics.AddError("Missing organization", "Organization must be set either on the resource or provider.")
		return
	}

	pkg := &client.Package{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		pkg.Description = plan.Description.ValueString()
	}

	tflog.Debug(ctx, "Creating package", map[string]interface{}{"org": org, "project": plan.Project.ValueString(), "name": pkg.Name})

	result, err := r.client.CreatePackage(org, plan.Project.ValueString(), pkg)
	if err != nil {
		resp.Diagnostics.AddError("Error creating package", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.Name = types.StringValue(result.Name)
	if result.Description != "" {
		plan.Description = types.StringValue(result.Description)
	}
	plan.LatestVersion = types.StringValue(result.LatestVersion)
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PackageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PackageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	result, err := r.client.GetPackage(org, state.Project.ValueString(), state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading package", err.Error())
		return
	}

	state.Organization = types.StringValue(org)
	state.Name = types.StringValue(result.Name)
	state.Description = types.StringValue(result.Description)
	state.LatestVersion = types.StringValue(result.LatestVersion)
	state.CreatedAt = types.StringValue(result.CreatedAt)
	state.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PackageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PackageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	pkg := &client.Package{}
	if !plan.Description.IsNull() {
		pkg.Description = plan.Description.ValueString()
	}

	result, err := r.client.UpdatePackage(org, plan.Project.ValueString(), plan.Name.ValueString(), pkg)
	if err != nil {
		resp.Diagnostics.AddError("Error updating package", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.Name = types.StringValue(result.Name)
	plan.Description = types.StringValue(result.Description)
	plan.LatestVersion = types.StringValue(result.LatestVersion)
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PackageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PackageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.DeletionProtection.ValueBool() {
		resp.Diagnostics.AddError(
			"Package is protected",
			fmt.Sprintf("Package %q has deletion_protection = true. Set it to false before destroying.", state.Name.ValueString()),
		)
		return
	}

	org := r.getOrg(&state)
	err := r.client.DeletePackage(org, state.Project.ValueString(), state.Name.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting package", err.Error())
	}
}

func (r *PackageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", "Import ID must be in the format: organization/project/package")
		return
	}

	org, project, name := parts[0], parts[1], parts[2]
	result, err := r.client.GetPackage(org, project, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing package", err.Error())
		return
	}

	state := PackageResourceModel{
		Organization:  types.StringValue(org),
		Project:       types.StringValue(project),
		Name:          types.StringValue(result.Name),
		Description:   types.StringValue(result.Description),
		LatestVersion: types.StringValue(result.LatestVersion),
		CreatedAt:     types.StringValue(result.CreatedAt),
		UpdatedAt:     types.StringValue(result.UpdatedAt),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
