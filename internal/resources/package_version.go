package resources

import (
	"archive/zip"
	"context"
	"fmt"
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
	"io"
	"os"
	"path/filepath"
)

var _ resource.Resource = &PackageVersionResource{}

type PackageVersionResource struct {
	client *client.Client
}

type PackageVersionResourceModel struct {
	Organization  types.String `tfsdk:"organization"`
	Environment   types.String `tfsdk:"environment"`
	PackageName   types.String `tfsdk:"package_name"`
	VersionID     types.String `tfsdk:"version_id"`
	SourceDir     types.String `tfsdk:"source_dir"`
	SourceFile    types.String `tfsdk:"source_file"`
	SourceHash    types.String `tfsdk:"source_hash"`
	Description   types.String `tfsdk:"description"`
	ArchiveStatus types.String `tfsdk:"archive_status"`
	BuildStatus   types.String `tfsdk:"build_status"`
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
		Description: "Publishes a version of a Credible package, which is also how a package is created -- the Admin API has no metadata-only create. Uploads either a local directory or a pre-built zip archive.",
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
				Description: "Path to a local directory. The provider zips its contents for upload.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("source_file")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_file": schema.StringAttribute{
				Description: "Path to a pre-built .zip archive. The API rejects other archive formats.",
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
			// Config-only, like source_hash: it is sent by the publish and never
			// read back, so it carries no plan modifier. Deliberately not
			// RequiresReplace -- a replace would archive this version and re-publish
			// the same version_id, which the API rejects with a 409, leaving no state
			// and an archived version. An edit is therefore absorbed by Update
			// without an API call, which matches the API: a package PATCH ignores
			// description.
			"description": schema.StringAttribute{
				Description: "Description applied to the package by this publish. Publishing is the only " +
					"operation that creates a package, so the description travels with the version. It " +
					"describes the package, not the version, and is only ever sent by the publish: editing " +
					"it later changes Terraform state without changing the package.",
				Optional: true,
			},
			"archive_status": schema.StringAttribute{
				Description: "Archive status: 'unarchive' (active) or 'archive' (archived).",
				Optional:    true,
				Computed:    true,
			},
			"build_status": schema.StringAttribute{
				Description: "Build status of the version, e.g. READY.",
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

// createArchiveFromDir zips a directory's contents and returns the temp file path.
// It must be a zip: the API reads the archive's bytes and rejects a tar.gz with a
// 500 regardless of the content type it is declared as.
func createArchiveFromDir(srcDir string) (string, error) {
	tmpFile, err := os.CreateTemp("", "credible-pkg-*.zip")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer tmpFile.Close()

	zipWriter := zip.NewWriter(tmpFile)

	srcDir = filepath.Clean(srcDir)

	err = filepath.Walk(srcDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Paths in the archive are relative to srcDir, so the package's files sit
		// at the archive root rather than under a copy of the local directory tree.
		relPath, err := filepath.Rel(srcDir, filePath)
		if err != nil {
			return fmt.Errorf("computing relative path: %w", err)
		}
		if relPath == "." {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("creating zip header for %s: %w", filePath, err)
		}
		header.Name = filepath.ToSlash(relPath)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("writing zip header for %s: %w", filePath, err)
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("opening file %s: %w", filePath, err)
		}
		defer file.Close()

		if _, err := io.Copy(writer, file); err != nil {
			return fmt.Errorf("writing file %s to zip: %w", filePath, err)
		}

		return nil
	})

	if err != nil {
		zipWriter.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("walking source directory: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("finalizing zip archive: %w", err)
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

	tflog.Debug(ctx, "Publishing package version", map[string]interface{}{
		"org": org, "environment": plan.Environment.ValueString(),
		"package": plan.PackageName.ValueString(), "version": version.ID,
	})

	if _, err := r.client.PublishPackageVersion(org, plan.Environment.ValueString(), plan.PackageName.ValueString(),
		plan.Description.ValueString(), version, uploadPath); err != nil {
		resp.Diagnostics.AddError("Error publishing package version", err.Error())
		return
	}

	// The publish returns the package, so the version's own fields come from a
	// read-back. It reads the version that was just requested rather than the
	// package's latestVersion, which names the currently *promoted* version and
	// so still points at an older version while this one is building.
	published, err := r.client.GetVersion(org, plan.Environment.ValueString(), plan.PackageName.ValueString(), version.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading published package version", err.Error())
		return
	}

	plan.Organization = types.StringValue(org)
	plan.VersionID = types.StringValue(published.ID)
	plan.ArchiveStatus = types.StringValue(published.ArchiveStatus)
	plan.BuildStatus = types.StringValue(published.BuildStatus)
	plan.CreatedAt = types.StringValue(published.CreatedAt)
	plan.UpdatedAt = types.StringValue(published.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PackageVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PackageVersionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := r.getOrg(&state)
	result, err := r.client.GetVersion(org, state.Environment.ValueString(), state.PackageName.ValueString(), state.VersionID.ValueString())
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
	state.BuildStatus = types.StringValue(result.BuildStatus)
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

		result, err := r.client.UpdateVersion(org, plan.Environment.ValueString(), plan.PackageName.ValueString(), plan.VersionID.ValueString(), version)
		if err != nil {
			resp.Diagnostics.AddError("Error updating package version", err.Error())
			return
		}

		plan.ArchiveStatus = types.StringValue(result.ArchiveStatus)
		plan.BuildStatus = types.StringValue(result.BuildStatus)
		plan.UpdatedAt = types.StringValue(result.UpdatedAt)
	}

	plan.Organization = types.StringValue(org)

	// description needs no request: it belongs to the package and the API ignores
	// it in a package PATCH, so the planned value is taken into state as-is.
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

	_, err := r.client.UpdateVersion(org, state.Environment.ValueString(), state.PackageName.ValueString(), state.VersionID.ValueString(), version)
	if err != nil && !client.IsNotFound(err) {
		// Log warning but don't fail — the version may already be archived or deleted
		tflog.Warn(ctx, "Could not archive version during delete", map[string]interface{}{"error": err.Error()})
	}
}
