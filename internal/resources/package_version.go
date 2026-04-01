package resources

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"github.com/credibledata/terraform-provider-credible/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &PackageVersionResource{}

type PackageVersionResource struct {
	client *client.Client
}

type PackageVersionResourceModel struct {
	Organization  types.String `tfsdk:"organization"`
	Project       types.String `tfsdk:"project"`
	PackageName   types.String `tfsdk:"package_name"`
	VersionID     types.String `tfsdk:"version_id"`
	SourceDir     types.String `tfsdk:"source_dir"`
	SourceFile    types.String `tfsdk:"source_file"`
	SourceHash    types.String `tfsdk:"source_hash"`
	ArchiveStatus types.String `tfsdk:"archive_status"`
	IndexStatus   types.String `tfsdk:"index_status"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

func NewPackageVersionResource() resource.Resource {
	return &PackageVersionResource{}
}

func (r *PackageVersionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package_version"
}

func (r *PackageVersionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Publishes a version of a Credible package. Supports uploading from a local directory or pre-built archive.",
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
			"package_name": schema.StringAttribute{
				Description: "The package name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version_id": schema.StringAttribute{
				Description: "The semantic version identifier (e.g., 1.0.0).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_dir": schema.StringAttribute{
				Description: "Path to a local directory. The provider will create a .tar.gz archive from its contents.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("source_file")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_file": schema.StringAttribute{
				Description: "Path to a pre-built .tar.gz archive.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("source_dir")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_hash": schema.StringAttribute{
				Description: "Hash of the source content, used for change detection. Use filemd5() for source_file.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"archive_status": schema.StringAttribute{
				Description: "Archive status: 'unarchive' (active) or 'archive' (archived).",
				Optional:    true,
				Computed:    true,
			},
			"index_status": schema.StringAttribute{
				Description: "Current indexing status of the version.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "When the version was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "When the version was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *PackageVersionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PackageVersionResource) getOrg(model *PackageVersionResourceModel) string {
	if !model.Organization.IsNull() && !model.Organization.IsUnknown() {
		return model.Organization.ValueString()
	}
	return r.client.Organization
}

// createArchiveFromDir creates a tar.gz archive from a directory and returns the temp file path.
func createArchiveFromDir(srcDir string) (string, error) {
	tmpFile, err := os.CreateTemp("", "credible-pkg-*.tar.gz")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer tmpFile.Close()

	gzWriter := gzip.NewWriter(tmpFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	srcDir = filepath.Clean(srcDir)

	err = filepath.Walk(srcDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("creating tar header for %s: %w", filePath, err)
		}

		// Use relative path within the archive
		relPath, err := filepath.Rel(srcDir, filePath)
		if err != nil {
			return fmt.Errorf("computing relative path: %w", err)
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("writing tar header: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("opening file %s: %w", filePath, err)
		}
		defer file.Close()

		if _, err := io.Copy(tarWriter, file); err != nil {
			return fmt.Errorf("writing file %s to tar: %w", filePath, err)
		}

		return nil
	})

	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("walking source directory: %w", err)
	}

	return tmpFile.Name(), nil
}

func (r *PackageVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PackageVersionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)
	if org == "" {
		resp.Diagnostics.AddError("Missing organization", "Organization must be set either on the resource or provider.")
		return
	}

	// Determine the file to upload
	var uploadPath string
	var tempFile string

	if !plan.SourceDir.IsNull() && plan.SourceDir.ValueString() != "" {
		// Create archive from directory
		var err error
		uploadPath, err = createArchiveFromDir(plan.SourceDir.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error creating archive from source_dir", err.Error())
			return
		}
		tempFile = uploadPath // Remember to clean up
	} else if !plan.SourceFile.IsNull() && plan.SourceFile.ValueString() != "" {
		uploadPath = plan.SourceFile.ValueString()
	} else {
		resp.Diagnostics.AddError("Missing source", "Either 'source_dir' or 'source_file' must be specified.")
		return
	}

	// Clean up temp file when done
	if tempFile != "" {
		defer os.Remove(tempFile)
	}

	version := &client.Version{
		ID: plan.VersionID.ValueString(),
	}

	tflog.Debug(ctx, "Creating package version", map[string]interface{}{
		"org": org, "project": plan.Project.ValueString(),
		"package": plan.PackageName.ValueString(), "version": version.ID,
	})

	result, err := r.client.CreateVersion(org, plan.Project.ValueString(), plan.PackageName.ValueString(), version, uploadPath)
	if err != nil {
		resp.Diagnostics.AddError("Error creating package version", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.VersionID = types.StringValue(result.ID)
	plan.ArchiveStatus = types.StringValue(result.ArchiveStatus)
	plan.IndexStatus = types.StringValue(result.IndexStatus)
	plan.CreatedAt = types.StringValue(result.CreatedAt)
	plan.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PackageVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PackageVersionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	result, err := r.client.GetVersion(org, state.Project.ValueString(), state.PackageName.ValueString(), state.VersionID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading package version", err.Error())
		return
	}

	state.Organization = types.StringValue(org)
	state.VersionID = types.StringValue(result.ID)
	state.ArchiveStatus = types.StringValue(result.ArchiveStatus)
	state.IndexStatus = types.StringValue(result.IndexStatus)
	state.CreatedAt = types.StringValue(result.CreatedAt)
	state.UpdatedAt = types.StringValue(result.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PackageVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PackageVersionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&plan)

	// Only archive_status can be updated on an existing version
	if !plan.ArchiveStatus.IsNull() && !plan.ArchiveStatus.IsUnknown() {
		version := &client.Version{
			ArchiveStatus: plan.ArchiveStatus.ValueString(),
		}

		result, err := r.client.UpdateVersion(org, plan.Project.ValueString(), plan.PackageName.ValueString(), plan.VersionID.ValueString(), version)
		if err != nil {
			resp.Diagnostics.AddError("Error updating package version", err.Error())
			return
		}

		plan.ArchiveStatus = types.StringValue(result.ArchiveStatus)
		plan.IndexStatus = types.StringValue(result.IndexStatus)
		plan.UpdatedAt = types.StringValue(result.UpdatedAt)
	}

	plan.Organization = types.StringValue(org)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PackageVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PackageVersionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Archive the version instead of deleting (versions are immutable)
	org := r.getOrg(&state)
	version := &client.Version{
		ArchiveStatus: "archive",
	}

	_, err := r.client.UpdateVersion(org, state.Project.ValueString(), state.PackageName.ValueString(), state.VersionID.ValueString(), version)
	if err != nil && !client.IsNotFound(err) {
		// Log warning but don't fail — the version may already be archived or deleted
		tflog.Warn(ctx, "Could not archive version during delete", map[string]interface{}{"error": err.Error()})
	}
}

