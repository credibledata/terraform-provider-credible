package resources

import (
	"context"
	"testing"

	"github.com/credibledata/terraform-provider-credible/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// These are plain unit tests -- no TF_ACC, no live API -- so they run in CI where
// the acceptance tests in this package self-skip. They pin the contract that every
// engine block reaches modelToAPI: adding a block to the schema without wiring it
// here is the failure mode these catch.

func str2(s string) types.String { return types.StringValue(s) }

func TestModelToAPI_EveryEngineIsWired(t *testing.T) {
	r := &ConnectionResource{}
	ctx := context.Background()

	tests := []struct {
		name  string
		model ConnectionResourceModel
		check func(t *testing.T, got *client.Connection)
	}{
		{
			name: "motherduck",
			model: ConnectionResourceModel{
				Name: str2("md"), Type: str2("motherduck"),
				Motherduck: &MotherduckModel{AccessToken: str2("tok"), Database: str2("analytics")},
			},
			check: func(t *testing.T, got *client.Connection) {
				if got.MotherDuckConnection == nil {
					t.Fatal("MotherDuckConnection not populated")
				}
				if got.MotherDuckConnection.AccessToken != "tok" {
					t.Errorf("AccessToken: expected %q, got %q", "tok", got.MotherDuckConnection.AccessToken)
				}
				if got.MotherDuckConnection.Database != "analytics" {
					t.Errorf("Database: expected %q, got %q", "analytics", got.MotherDuckConnection.Database)
				}
			},
		},
		{
			name: "databricks",
			model: ConnectionResourceModel{
				Name: str2("dbx"), Type: str2("databricks"),
				Databricks: &DatabricksModel{Host: str2("h"), Path: str2("/p"), Token: str2("t"), DefaultCatalog: str2("main")},
			},
			check: func(t *testing.T, got *client.Connection) {
				if got.DatabricksConnection == nil {
					t.Fatal("DatabricksConnection not populated")
				}
				if got.DatabricksConnection.Token != "t" {
					t.Errorf("Token: expected %q, got %q", "t", got.DatabricksConnection.Token)
				}
				if got.DatabricksConnection.DefaultCatalog != "main" {
					t.Errorf("DefaultCatalog: expected %q, got %q", "main", got.DatabricksConnection.DefaultCatalog)
				}
			},
		},
		{
			name: "publisher",
			model: ConnectionResourceModel{
				Name: str2("pub"), Type: str2("publisher"),
				Publisher: &PublisherModel{ConnectionUri: str2("https://x/api/v0/y"), AccessToken: str2("tok")},
			},
			check: func(t *testing.T, got *client.Connection) {
				if got.PublisherConnection == nil {
					t.Fatal("PublisherConnection not populated")
				}
				if got.PublisherConnection.ConnectionUri != "https://x/api/v0/y" {
					t.Errorf("ConnectionUri: got %q", got.PublisherConnection.ConnectionUri)
				}
			},
		},
		{
			name: "duckdb with attached database",
			model: ConnectionResourceModel{
				Name: str2("dd"), Type: str2("duckdb"),
				Duckdb: &DuckdbModel{AttachedDatabases: []AttachedDatabaseModel{{
					Name: str2("wh"), Type: str2("s3"),
					S3: &S3Model{Provider: str2("config"), AccessKeyId: str2("AKIA"), SecretAccessKey: str2("shh")},
				}}},
			},
			check: func(t *testing.T, got *client.Connection) {
				if got.DuckdbConnection == nil {
					t.Fatal("DuckdbConnection not populated")
				}
				if len(got.DuckdbConnection.AttachedDatabases) != 1 {
					t.Fatalf("expected 1 attached database, got %d", len(got.DuckdbConnection.AttachedDatabases))
				}
				ad := got.DuckdbConnection.AttachedDatabases[0]
				if ad.Name != "wh" || ad.Type != "s3" {
					t.Errorf("attached database identity: got name=%q type=%q", ad.Name, ad.Type)
				}
				if ad.S3Connection == nil || ad.S3Connection.SecretAccessKey != "shh" {
					t.Errorf("S3 secret not carried through: %+v", ad.S3Connection)
				}
			},
		},
		{
			name: "ducklake",
			model: ConnectionResourceModel{
				Name: str2("dl"), Type: str2("ducklake"),
				Ducklake: &DucklakeModel{
					Storage: &DucklakeStorageModel{BucketUrl: str2("s3://b/p"), Gcs: &GcsModel{KeyId: str2("k"), Secret: str2("s")}},
					Catalog: &DucklakeCatalogModel{Postgres: &PostgresModel{Host: str2("pg")}, MetadataSchema: str2("meta")},
				},
			},
			check: func(t *testing.T, got *client.Connection) {
				if got.DucklakeConnection == nil {
					t.Fatal("DucklakeConnection not populated")
				}
				if got.DucklakeConnection.Storage == nil || got.DucklakeConnection.Storage.BucketUrl != "s3://b/p" {
					t.Errorf("storage: %+v", got.DucklakeConnection.Storage)
				}
				if got.DucklakeConnection.Catalog == nil || got.DucklakeConnection.Catalog.MetadataSchema != "meta" {
					t.Errorf("catalog: %+v", got.DucklakeConnection.Catalog)
				}
				if got.DucklakeConnection.Catalog.PostgresConnection == nil {
					t.Error("catalog postgres connection not populated")
				}
			},
		},
		{
			name: "postgres with ssh proxy",
			model: ConnectionResourceModel{
				Name: str2("pg"), Type: str2("postgres"),
				Postgres: &PostgresModel{Host: str2("db.internal"), Sslmode: str2("verify-full")},
				Proxy: &ProxyModel{Type: str2("ssh"), Ssh: &SshModel{
					Host: str2("bastion"), Port: types.Int64Value(2222), Username: str2("tf"), PrivateKey: str2("PEM"),
				}},
			},
			check: func(t *testing.T, got *client.Connection) {
				if got.Proxy == nil || got.Proxy.Ssh == nil {
					t.Fatal("Proxy/Ssh not populated")
				}
				if got.Proxy.Ssh.Port == nil || *got.Proxy.Ssh.Port != 2222 {
					t.Errorf("ssh port: got %v", got.Proxy.Ssh.Port)
				}
				if got.Proxy.Ssh.PrivateKey != "PEM" {
					t.Errorf("private key: got %q", got.Proxy.Ssh.PrivateKey)
				}
				if got.PostgresConnection == nil || got.PostgresConnection.Sslmode != "verify-full" {
					t.Errorf("sslmode: %+v", got.PostgresConnection)
				}
			},
		},
		{
			name: "trino credentials",
			model: ConnectionResourceModel{
				Name: str2("tr"), Type: str2("trino"),
				Trino: &TrinoModel{Server: str2("s"), Password: str2("pw"), PeakaKey: str2("pk")},
			},
			check: func(t *testing.T, got *client.Connection) {
				if got.TrinoConnection == nil {
					t.Fatal("TrinoConnection not populated")
				}
				if got.TrinoConnection.Password != "pw" || got.TrinoConnection.PeakaKey != "pk" {
					t.Errorf("trino credentials: %+v", got.TrinoConnection)
				}
			},
		},
		{
			name: "bigquery impersonation",
			model: ConnectionResourceModel{
				Name: str2("bq"), Type: str2("bigquery"),
				Bigquery: &BigqueryModel{ImpersonateServiceAccount: str2("sa@p.iam.gserviceaccount.com")},
			},
			check: func(t *testing.T, got *client.Connection) {
				if got.BigqueryConnection == nil {
					t.Fatal("BigqueryConnection not populated")
				}
				if got.BigqueryConnection.ImpersonateServiceAccount != "sa@p.iam.gserviceaccount.com" {
					t.Errorf("impersonation: got %q", got.BigqueryConnection.ImpersonateServiceAccount)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.check(t, r.modelToAPI(ctx, &tc.model))
		})
	}
}

// A block the practitioner did not configure must stay nil, so `omitempty` keeps it
// off the wire. Sending an empty object changes API behavior (a present `proxy`
// makes the connection proxied).
func TestModelToAPI_UnsetBlocksStayNil(t *testing.T) {
	r := &ConnectionResource{}
	got := r.modelToAPI(context.Background(), &ConnectionResourceModel{
		Name: str2("plain"), Type: str2("postgres"),
		Postgres: &PostgresModel{Host: str2("db")},
	})

	if got.Proxy != nil {
		t.Error("Proxy must stay nil when no proxy block is configured")
	}
	if got.DuckdbConnection != nil || got.MotherDuckConnection != nil ||
		got.DucklakeConnection != nil || got.PublisherConnection != nil || got.DatabricksConnection != nil {
		t.Error("engine blocks other than the configured one must stay nil")
	}
}

// The API never returns secrets, so apiToModel must restore them from prior state.
// Without this the attribute reads back null and Terraform plans a permanent diff.
//
// The response below carries each block structurally present with blank secrets --
// exactly what a real GET returns -- so the test exercises the rebuild path rather
// than a shape production never produces.
func TestApiToModel_PreservesSecretsTheAPIOmits(t *testing.T) {
	r := &ConnectionResource{}
	ctx := context.Background()

	prior := &ConnectionResourceModel{
		Motherduck: &MotherduckModel{AccessToken: str2("md-secret")},
		Databricks: &DatabricksModel{Token: str2("dbx-token"), OauthClientSecret: str2("dbx-oauth")},
		Publisher:  &PublisherModel{AccessToken: str2("pub-token")},
		Trino:      &TrinoModel{Password: str2("trino-pw"), PeakaKey: str2("peaka")},
		Postgres:   &PostgresModel{Password: str2("pg-pw")},
		Proxy:      &ProxyModel{Ssh: &SshModel{PrivateKey: str2("PEM"), PrivateKeyPass: str2("pass")}},
		Ducklake: &DucklakeModel{
			Storage: &DucklakeStorageModel{Gcs: &GcsModel{Secret: str2("gcs-secret")}},
			Catalog: &DucklakeCatalogModel{Postgres: &PostgresModel{Password: str2("cat-pw")}},
		},
		Duckdb: &DuckdbModel{AttachedDatabases: []AttachedDatabaseModel{
			{Name: str2("wh"), S3: &S3Model{SecretAccessKey: str2("s3-secret")}},
		}},
	}

	// A response as the API actually sends it: blocks present, secrets blank.
	result := &client.Connection{
		Name: "c", Type: "postgres",
		PostgresConnection:   &client.PostgresConnection{Host: "db"},
		TrinoConnection:      &client.TrinoConnection{Server: "s"},
		DatabricksConnection: &client.DatabricksConnection{Host: "h"},
		MotherDuckConnection: &client.MotherDuckConnection{Database: "d"},
		PublisherConnection:  &client.PublisherConnection{ConnectionUri: "u"},
		Proxy:                &client.ConnectionProxy{Type: "ssh", Ssh: &client.SshProxyConfig{Host: "bastion"}},
		DucklakeConnection: &client.DucklakeConnection{
			Storage: &client.DucklakeStorage{BucketUrl: "s3://b", GcsConnection: &client.GCSConnection{KeyId: "k"}},
			Catalog: &client.DucklakeCatalog{PostgresConnection: &client.PostgresConnection{Host: "pg"}},
		},
		DuckdbConnection: &client.DuckdbConnection{AttachedDatabases: []client.AttachedDatabase{
			{Name: "wh", Type: "s3", S3Connection: &client.S3Connection{AccessKeyId: "AKIA"}},
		}},
	}

	model := &ConnectionResourceModel{}
	r.apiToModel(ctx, result, model, prior)

	for _, tc := range []struct {
		field string
		got   types.String
		want  string
	}{
		{"postgres.password", model.Postgres.Password, "pg-pw"},
		{"motherduck.access_token", model.Motherduck.AccessToken, "md-secret"},
		{"databricks.token", model.Databricks.Token, "dbx-token"},
		{"databricks.oauth_client_secret", model.Databricks.OauthClientSecret, "dbx-oauth"},
		{"publisher.access_token", model.Publisher.AccessToken, "pub-token"},
		{"trino.password", model.Trino.Password, "trino-pw"},
		{"trino.peaka_key", model.Trino.PeakaKey, "peaka"},
		{"proxy.ssh.private_key", model.Proxy.Ssh.PrivateKey, "PEM"},
		{"proxy.ssh.private_key_pass", model.Proxy.Ssh.PrivateKeyPass, "pass"},
		{"ducklake.storage.gcs.secret", model.Ducklake.Storage.Gcs.Secret, "gcs-secret"},
		{"ducklake.catalog.postgres.password", model.Ducklake.Catalog.Postgres.Password, "cat-pw"},
		{"duckdb.attached_databases[0].s3.secret_access_key", model.Duckdb.AttachedDatabases[0].S3.SecretAccessKey, "s3-secret"},
	} {
		if tc.got.ValueString() != tc.want {
			t.Errorf("%s: expected %q to be preserved, got %q", tc.field, tc.want, tc.got.ValueString())
		}
	}

	// Non-secret fields must come from the response, not prior state.
	if model.Postgres.Host.ValueString() != "db" {
		t.Errorf("host must be refreshed from the API, got %q", model.Postgres.Host.ValueString())
	}
}

// Create and Update pass the same model as both source and preserve arguments.
// A shallow copy would alias the pointers being overwritten and wipe the secrets.
func TestApiToModel_SelfAliasedCallKeepsSecrets(t *testing.T) {
	r := &ConnectionResource{}
	plan := &ConnectionResourceModel{
		Postgres: &PostgresModel{Host: str2("db"), Password: str2("pg-pw")},
		Proxy:    &ProxyModel{Ssh: &SshModel{PrivateKey: str2("PEM")}},
	}

	r.apiToModel(context.Background(), &client.Connection{
		Name: "c", Type: "postgres",
		PostgresConnection: &client.PostgresConnection{Host: "db"},
		Proxy:              &client.ConnectionProxy{Type: "ssh", Ssh: &client.SshProxyConfig{Host: "bastion"}},
	}, plan, plan)

	if got := plan.Postgres.Password.ValueString(); got != "pg-pw" {
		t.Errorf("password lost on self-aliased call: got %q", got)
	}
	if got := plan.Proxy.Ssh.PrivateKey.ValueString(); got != "PEM" {
		t.Errorf("ssh private key lost on self-aliased call: got %q", got)
	}
}

// Secrets pair to an attachment by name. Index pairing would hand each remaining
// attachment its neighbour's credential once an earlier entry is removed.
func TestApiToModel_AttachedDatabaseSecretsPairByName(t *testing.T) {
	r := &ConnectionResource{}
	prior := &ConnectionResourceModel{Duckdb: &DuckdbModel{AttachedDatabases: []AttachedDatabaseModel{
		{Name: str2("alpha"), S3: &S3Model{SecretAccessKey: str2("secret-alpha")}},
		{Name: str2("beta"), S3: &S3Model{SecretAccessKey: str2("secret-beta")}},
	}}}

	// "alpha" was removed from config, so "beta" is now at index 0.
	model := &ConnectionResourceModel{}
	r.apiToModel(context.Background(), &client.Connection{
		Name: "c", Type: "duckdb",
		DuckdbConnection: &client.DuckdbConnection{AttachedDatabases: []client.AttachedDatabase{
			{Name: "beta", Type: "s3", S3Connection: &client.S3Connection{AccessKeyId: "AKIA"}},
		}},
	}, model, prior)

	got := model.Duckdb.AttachedDatabases[0].S3.SecretAccessKey.ValueString()
	if got != "secret-beta" {
		t.Errorf("expected beta to keep its own secret, got %q (index pairing would give secret-alpha)", got)
	}
}

// An engine changed out of band must not leave the previous block behind: that
// state is unreachable from any config and the next apply would send the wrong
// engine's payload.
func TestApiToModel_DropsBlocksTheResponseNoLongerCarries(t *testing.T) {
	r := &ConnectionResource{}
	model := &ConnectionResourceModel{Postgres: &PostgresModel{Host: str2("old")}}

	r.apiToModel(context.Background(), &client.Connection{
		Name: "c", Type: "mysql",
		MysqlConnection: &client.MysqlConnection{Host: "mysql-host"},
	}, model, nil)

	if model.Postgres != nil {
		t.Error("postgres block must be dropped when the response no longer carries it")
	}
	if model.Mysql == nil || model.Mysql.Host.ValueString() != "mysql-host" {
		t.Errorf("mysql block must be built from the response, got %+v", model.Mysql)
	}
}

// Import has no prior state, so the engine block must be rebuilt from the response
// or the next plan proposes adding a block that already exists.
func TestPopulateBlocksFromAPI_RebuildsEngineBlockOnImport(t *testing.T) {
	port := 5432
	model := &ConnectionResourceModel{}
	populateBlocksFromAPI(&client.Connection{
		Name: "c", Type: "postgres",
		PostgresConnection: &client.PostgresConnection{Host: "db", Port: &port, DatabaseName: "app"},
		Proxy:              &client.ConnectionProxy{Type: "ssh", Ssh: &client.SshProxyConfig{Host: "bastion"}},
	}, model)

	if model.Postgres == nil {
		t.Fatal("postgres block not reconstructed on import")
	}
	if model.Postgres.Host.ValueString() != "db" {
		t.Errorf("host: got %q", model.Postgres.Host.ValueString())
	}
	if model.Postgres.Port.ValueInt64() != 5432 {
		t.Errorf("port: got %d", model.Postgres.Port.ValueInt64())
	}
	if model.Proxy == nil || model.Proxy.Ssh == nil || model.Proxy.Ssh.Host.ValueString() != "bastion" {
		t.Errorf("proxy not reconstructed: %+v", model.Proxy)
	}
	// A field the API omitted must stay null, not become "".
	if !model.Postgres.Password.IsNull() {
		t.Error("a field absent from the response must be null, not an empty string")
	}
}
