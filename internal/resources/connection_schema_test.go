package resources

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Terraform redacts an attribute in plan and apply output only when the schema
// marks it Sensitive, and the API gives no signal of its own about which fields are
// credentials. These tests pin the mark on every credential the connection schema
// accepts, so an engine block added or edited without it fails here instead of
// printing the value in a plan.
//
// Entries are keyed by "<block>.<attribute>", where the block is the attribute's
// immediate parent. The postgres, bigquery, snowflake, gcs, s3 and azure blocks
// come from shared helpers and appear under several paths, so one entry covers
// every place a helper is used.

// connectionSecrets is every attribute that holds a credential. Each one must be
// Sensitive, and every Sensitive attribute must be listed here.
var connectionSecrets = map[string]struct{}{
	"postgres.password":                 {},
	"postgres.connection_string":        {},
	"bigquery.service_account_key_json": {},
	"snowflake.password":                {},
	"snowflake.private_key":             {},
	"snowflake.private_key_pass":        {},
	"trino.password":                    {},
	"trino.peaka_key":                   {},
	"databricks.token":                  {},
	"databricks.oauth_client_secret":    {},
	"mysql.password":                    {},
	"motherduck.access_token":           {},
	"publisher.access_token":            {},
	"ssh.private_key":                   {},
	"ssh.private_key_pass":              {},
	"gcs.secret":                        {},
	"s3.secret_access_key":              {},
	"s3.session_token":                  {},
	"azure.sas_url":                     {},
	"azure.client_secret":               {},
}

// connectionNotSecrets is every attribute whose name looks like a credential but
// holds an identifier or public material, with the reason it stays unmarked.
var connectionNotSecrets = map[string]string{
	"gcs.key_id":       "the HMAC access key ID, which identifies the key; the secret is gcs.secret",
	"s3.access_key_id": "the AWS access key ID, which identifies the key; the secret is s3.secret_access_key",
	"ssh.host_key":     "the bastion's public host key, pinned for verification",
}

// credentialNameParts are name fragments that suggest an attribute holds a
// credential. An attribute whose name contains one must appear in
// connectionSecrets or connectionNotSecrets.
var credentialNameParts = []string{
	"password", "pass", "secret", "token", "key", "sas", "credential", "connection_string",
}

type connectionSchemaAttr struct {
	path      string // full dotted path, for messages
	key       string // "<block>.<attribute>", or the bare name at the top level
	sensitive bool
}

func connectionSchemaAttributes(t *testing.T) []connectionSchemaAttr {
	t.Helper()
	resp := &resource.SchemaResponse{}
	(&ConnectionResource{}).Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("connection schema returned diagnostics: %v", resp.Diagnostics)
	}

	var out []connectionSchemaAttr
	walkSchemaAttributes(t, "", "", resp.Schema.Attributes, &out)
	walkSchemaBlocks(t, "", resp.Schema.Blocks, &out)
	sort.Slice(out, func(i, j int) bool { return out[i].path < out[j].path })
	return out
}

func walkSchemaAttributes(t *testing.T, prefix, parent string, attrs map[string]schema.Attribute, out *[]connectionSchemaAttr) {
	t.Helper()
	for name, attr := range attrs {
		path := prefix + name
		// A nested attribute carries attributes of its own, which this walk would
		// otherwise skip without saying so.
		if _, nested := attr.(schema.NestedAttribute); nested {
			t.Fatalf("%s is a nested attribute; extend walkSchemaAttributes to descend into it", path)
		}
		key := name
		if parent != "" {
			key = parent + "." + name
		}
		*out = append(*out, connectionSchemaAttr{path: path, key: key, sensitive: attr.IsSensitive()})
	}
}

func walkSchemaBlocks(t *testing.T, prefix string, blocks map[string]schema.Block, out *[]connectionSchemaAttr) {
	t.Helper()
	for name, block := range blocks {
		path := prefix + name + "."
		switch b := block.(type) {
		case schema.SingleNestedBlock:
			walkSchemaAttributes(t, path, name, b.Attributes, out)
			walkSchemaBlocks(t, path, b.Blocks, out)
		case schema.ListNestedBlock:
			walkSchemaAttributes(t, path, name, b.NestedObject.Attributes, out)
			walkSchemaBlocks(t, path, b.NestedObject.Blocks, out)
		case schema.SetNestedBlock:
			walkSchemaAttributes(t, path, name, b.NestedObject.Attributes, out)
			walkSchemaBlocks(t, path, b.NestedObject.Blocks, out)
		default:
			t.Fatalf("%s is a %T; extend walkSchemaBlocks to descend into it", prefix+name, block)
		}
	}
}

func TestConnectionSchema_CredentialsAreSensitive(t *testing.T) {
	seen := map[string]bool{}
	for _, a := range connectionSchemaAttributes(t) {
		seen[a.key] = true
		_, secret := connectionSecrets[a.key]
		switch {
		case secret && !a.sensitive:
			t.Errorf("%s holds a credential but is not marked Sensitive", a.path)
		case !secret && a.sensitive:
			t.Errorf("%s is marked Sensitive but %q is missing from connectionSecrets", a.path, a.key)
		}
	}
	for key := range connectionSecrets {
		if !seen[key] {
			t.Errorf("connectionSecrets lists %q, but no attribute in the schema has that key", key)
		}
	}
}

func TestConnectionSchema_CredentialLikeNamesAreClassified(t *testing.T) {
	seen := map[string]bool{}
	for _, a := range connectionSchemaAttributes(t) {
		seen[a.key] = true
		if _, secret := connectionSecrets[a.key]; secret {
			continue
		}
		if _, cleared := connectionNotSecrets[a.key]; cleared {
			continue
		}
		name := a.key[strings.LastIndex(a.key, ".")+1:]
		for _, part := range credentialNameParts {
			if strings.Contains(name, part) {
				t.Errorf("%s has %q in its name but is in neither connectionSecrets nor connectionNotSecrets: "+
					"mark it Sensitive and list it as a secret, or list it as not secret with the reason", a.path, part)
				break
			}
		}
	}
	for key := range connectionNotSecrets {
		if !seen[key] {
			t.Errorf("connectionNotSecrets lists %q, but no attribute in the schema has that key", key)
		}
	}
}
