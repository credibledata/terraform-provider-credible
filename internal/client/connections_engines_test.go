package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The API never round-trips these structs through the provider's own decoder in
// production -- the server does. So these tests assert the MARSHALLED JSON rather
// than Go struct fields: a wrong `json:` tag (the exact defect that left duckdb and
// motherduck sending `md_token`/`url`, neither of which the API accepts) is
// invisible to a struct-level assertion but fails here.
func createConnectionAndCaptureBody(t *testing.T, conn *Connection) map[string]any {
	t.Helper()

	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
		}
		if err := json.Unmarshal(raw, &captured); err != nil {
			t.Errorf("request body is not valid JSON: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Connection{Name: conn.Name, Type: conn.Type})
	}))
	defer server.Close()

	c := NewClient(server.URL, "ApiKey k", "org")
	if _, err := c.CreateConnection("my-org", "my-env", conn); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("server never received a request body")
	}
	return captured
}

// nested walks a decoded JSON body, failing the test if any key is missing.
func nested(t *testing.T, body map[string]any, keys ...string) map[string]any {
	t.Helper()
	cur := body
	for _, k := range keys {
		next, ok := cur[k].(map[string]any)
		if !ok {
			t.Fatalf("expected object at %q, body was: %v", strings.Join(keys, "."), body)
		}
		cur = next
	}
	return cur
}

func TestConnectionWireShape_MotherDuck(t *testing.T) {
	body := createConnectionAndCaptureBody(t, &Connection{
		Name: "md", Type: "motherduck",
		MotherDuckConnection: &MotherDuckConnection{AccessToken: "tok", Database: "analytics"},
	})

	// Wrapper key keeps the API's lowercase "d" even though the schema is MotherDuckConnection.
	md := nested(t, body, "motherduckConnection")
	if got := md["accessToken"]; got != "tok" {
		t.Errorf("accessToken: expected %q, got %v", "tok", got)
	}
	if got := md["database"]; got != "analytics" {
		t.Errorf("database: expected %q, got %v", "analytics", got)
	}
	// The fields the provider used to send, which the API does not define.
	if _, present := md["md_token"]; present {
		t.Error("md_token must not be sent: the API has no such field")
	}
	if _, present := md["url"]; present {
		t.Error("url must not be sent: the API has no such field")
	}
}

func TestConnectionWireShape_Duckdb(t *testing.T) {
	body := createConnectionAndCaptureBody(t, &Connection{
		Name: "dd", Type: "duckdb",
		DuckdbConnection: &DuckdbConnection{
			AttachedDatabases: []AttachedDatabase{{
				Name: "warehouse", Type: "s3",
				S3Connection: &S3Connection{Provider: "config", AccessKeyId: "AKIA", SecretAccessKey: "shh", Region: "us-west-2"},
			}},
		},
	})

	dd := nested(t, body, "duckdbConnection")
	if _, present := dd["md_token"]; present {
		t.Error("md_token must not be sent: DuckDB has no token field")
	}
	if _, present := dd["url"]; present {
		t.Error("url must not be sent: DuckDB has no url field")
	}

	attached, ok := dd["attachedDatabases"].([]any)
	if !ok || len(attached) != 1 {
		t.Fatalf("expected 1 attachedDatabases entry, got %v", dd["attachedDatabases"])
	}
	entry, ok := attached[0].(map[string]any)
	if !ok {
		t.Fatalf("attachedDatabases[0] is not an object: %v", attached[0])
	}
	if got := entry["name"]; got != "warehouse" {
		t.Errorf("name: expected %q, got %v", "warehouse", got)
	}
	s3, ok := entry["s3Connection"].(map[string]any)
	if !ok {
		t.Fatalf("expected s3Connection object, got %v", entry["s3Connection"])
	}
	if got := s3["secretAccessKey"]; got != "shh" {
		t.Errorf("secretAccessKey: expected %q, got %v", "shh", got)
	}
	if got := s3["region"]; got != "us-west-2" {
		t.Errorf("region: expected %q, got %v", "us-west-2", got)
	}
}

func TestConnectionWireShape_Databricks(t *testing.T) {
	body := createConnectionAndCaptureBody(t, &Connection{
		Name: "dbx", Type: "databricks",
		DatabricksConnection: &DatabricksConnection{
			Host: "dbc-1.cloud.databricks.com", Path: "/sql/1.0/warehouses/abc",
			Token: "dapi", OauthClientId: "cid", OauthClientSecret: "csec",
			DefaultCatalog: "main", DefaultSchema: "public", SetupSQL: "SET x = 1",
		},
	})

	dbx := nested(t, body, "databricksConnection")
	for key, want := range map[string]string{
		"host":              "dbc-1.cloud.databricks.com",
		"path":              "/sql/1.0/warehouses/abc",
		"token":             "dapi",
		"oauthClientId":     "cid",
		"oauthClientSecret": "csec",
		"defaultCatalog":    "main",
		"defaultSchema":     "public",
		"setupSQL":          "SET x = 1",
	} {
		if got := dbx[key]; got != want {
			t.Errorf("%s: expected %q, got %v", key, want, got)
		}
	}
}

func TestConnectionWireShape_Publisher(t *testing.T) {
	body := createConnectionAndCaptureBody(t, &Connection{
		Name: "pub", Type: "publisher",
		PublisherConnection: &PublisherConnection{ConnectionUri: "https://example/api/v0/x", AccessToken: "tok"},
	})

	pub := nested(t, body, "publisherConnection")
	if got := pub["connectionUri"]; got != "https://example/api/v0/x" {
		t.Errorf("connectionUri: expected the configured URI, got %v", got)
	}
	if got := pub["accessToken"]; got != "tok" {
		t.Errorf("accessToken: expected %q, got %v", "tok", got)
	}
}

func TestConnectionWireShape_Ducklake(t *testing.T) {
	port := 5432
	body := createConnectionAndCaptureBody(t, &Connection{
		Name: "dl", Type: "ducklake",
		DucklakeConnection: &DucklakeConnection{
			Storage: &DucklakeStorage{
				BucketUrl:     "s3://bucket/prefix",
				GcsConnection: &GCSConnection{KeyId: "kid", Secret: "sec"},
			},
			Catalog: &DucklakeCatalog{
				PostgresConnection: &PostgresConnection{Host: "pg", Port: &port, DatabaseName: "cat"},
				MetadataSchema:     "ducklake_meta",
			},
		},
	})

	storage := nested(t, body, "ducklakeConnection", "storage")
	if got := storage["bucketUrl"]; got != "s3://bucket/prefix" {
		t.Errorf("bucketUrl: expected the configured URL, got %v", got)
	}
	gcs := nested(t, body, "ducklakeConnection", "storage", "gcsConnection")
	if got := gcs["secret"]; got != "sec" {
		t.Errorf("gcs secret: expected %q, got %v", "sec", got)
	}

	catalog := nested(t, body, "ducklakeConnection", "catalog")
	if got := catalog["metadataSchema"]; got != "ducklake_meta" {
		t.Errorf("metadataSchema: expected %q, got %v", "ducklake_meta", got)
	}
	pg := nested(t, body, "ducklakeConnection", "catalog", "postgresConnection")
	if got := pg["port"]; got != float64(5432) {
		t.Errorf("catalog postgres port: expected 5432, got %v", got)
	}
}

func TestConnectionWireShape_Proxy(t *testing.T) {
	port := 2222
	body := createConnectionAndCaptureBody(t, &Connection{
		Name: "pg-via-bastion", Type: "postgres",
		PostgresConnection: &PostgresConnection{Host: "db.internal", Sslmode: "verify-full"},
		Proxy: &ConnectionProxy{
			Type: "ssh",
			Ssh:  &SshProxyConfig{Host: "bastion", Port: &port, Username: "tf", PrivateKey: "PEM", HostKey: "ssh-ed25519 AAAA"},
		},
	})

	proxy := nested(t, body, "proxy")
	if got := proxy["type"]; got != "ssh" {
		t.Errorf("proxy type: expected %q, got %v", "ssh", got)
	}
	ssh := nested(t, body, "proxy", "ssh")
	if got := ssh["host"]; got != "bastion" {
		t.Errorf("ssh host: expected %q, got %v", "bastion", got)
	}
	if got := ssh["port"]; got != float64(2222) {
		t.Errorf("ssh port: expected 2222, got %v", got)
	}
	if got := ssh["privateKey"]; got != "PEM" {
		t.Errorf("privateKey: expected %q, got %v", "PEM", got)
	}

	// sslmode rides on the postgres object, not the proxy.
	pg := nested(t, body, "postgresConnection")
	if got := pg["sslmode"]; got != "verify-full" {
		t.Errorf("sslmode: expected %q, got %v", "verify-full", got)
	}
}

// An unset proxy must be absent entirely rather than sent as null: the API treats a
// present `proxy` key as "this connection is proxied" and then applies proxy-only
// validation (e.g. it rejects sslmode on a direct connection).
func TestConnectionWireShape_OmitsUnsetOptionalObjects(t *testing.T) {
	body := createConnectionAndCaptureBody(t, &Connection{
		Name: "plain", Type: "postgres",
		PostgresConnection: &PostgresConnection{Host: "db"},
	})

	for _, key := range []string{
		"proxy", "duckdbConnection", "motherduckConnection", "ducklakeConnection",
		"publisherConnection", "databricksConnection",
	} {
		if _, present := body[key]; present {
			t.Errorf("%s must be omitted when unset, but was present in the request body", key)
		}
	}
}

// Azure and Snowflake attachments, and the two fields added to existing engines,
// are only reachable through nested positions -- a mis-tagged property there is
// silently dropped by the server rather than rejected, so assert them on the wire.
func TestConnectionWireShape_AttachedAzureAndSnowflake(t *testing.T) {
	body := createConnectionAndCaptureBody(t, &Connection{
		Name: "dd", Type: "duckdb",
		DuckdbConnection: &DuckdbConnection{AttachedDatabases: []AttachedDatabase{
			{
				Name: "blob", Type: "azure",
				AzureConnection: &AzureConnection{
					AuthType: "service_principal", TenantId: "tid", ClientId: "cid",
					ClientSecret: "csec", AccountName: "acct", FileUrl: "https://a/b/**",
					SasUrl: "https://a/c?sp=rl",
				},
			},
			{
				Name: "sf", Type: "snowflake",
				SnowflakeConnection: &SnowflakeConnection{Account: "acc", Username: "u", PrivateKey: "PEM", Warehouse: "wh"},
			},
		}},
	})

	attached, ok := nested(t, body, "duckdbConnection")["attachedDatabases"].([]any)
	if !ok || len(attached) != 2 {
		t.Fatalf("expected 2 attachments, got %v", attached)
	}

	az, ok := attached[0].(map[string]any)["azureConnection"].(map[string]any)
	if !ok {
		t.Fatalf("azureConnection missing: %v", attached[0])
	}
	for key, want := range map[string]string{
		"authType": "service_principal", "tenantId": "tid", "clientId": "cid",
		"clientSecret": "csec", "accountName": "acct", "fileUrl": "https://a/b/**",
		"sasUrl": "https://a/c?sp=rl",
	} {
		if got := az[key]; got != want {
			t.Errorf("azure %s: expected %q, got %v", key, want, got)
		}
	}

	sf, ok := attached[1].(map[string]any)["snowflakeConnection"].(map[string]any)
	if !ok {
		t.Fatalf("snowflakeConnection missing: %v", attached[1])
	}
	if got := sf["privateKey"]; got != "PEM" {
		t.Errorf("snowflake privateKey: expected %q, got %v", "PEM", got)
	}
	if got := sf["account"]; got != "acc" {
		t.Errorf("snowflake account: expected %q, got %v", "acc", got)
	}
}

func TestConnectionWireShape_NewFieldsOnExistingEngines(t *testing.T) {
	body := createConnectionAndCaptureBody(t, &Connection{
		Name: "bq", Type: "bigquery",
		BigqueryConnection: &BigqueryConnection{ImpersonateServiceAccount: "sa@p.iam.gserviceaccount.com"},
	})
	if got := nested(t, body, "bigqueryConnection")["impersonateServiceAccount"]; got != "sa@p.iam.gserviceaccount.com" {
		t.Errorf("impersonateServiceAccount: got %v", got)
	}

	body = createConnectionAndCaptureBody(t, &Connection{
		Name: "tr", Type: "trino",
		TrinoConnection: &TrinoConnection{Server: "s", Password: "pw", PeakaKey: "pk"},
	})
	tr := nested(t, body, "trinoConnection")
	if got := tr["password"]; got != "pw" {
		t.Errorf("trino password: got %v", got)
	}
	if got := tr["peakaKey"]; got != "pk" {
		t.Errorf("trino peakaKey: got %v", got)
	}

	body = createConnectionAndCaptureBody(t, &Connection{
		Name: "s3", Type: "duckdb",
		DuckdbConnection: &DuckdbConnection{AttachedDatabases: []AttachedDatabase{{
			Name: "chain", Type: "s3",
			S3Connection: &S3Connection{Provider: "credential_chain", Chain: "env;instance", SessionToken: "tok", Endpoint: "https://minio"},
		}}},
	})
	attached := nested(t, body, "duckdbConnection")["attachedDatabases"].([]any)
	s3 := attached[0].(map[string]any)["s3Connection"].(map[string]any)
	for key, want := range map[string]string{
		"provider": "credential_chain", "chain": "env;instance",
		"sessionToken": "tok", "endpoint": "https://minio",
	} {
		if got := s3[key]; got != want {
			t.Errorf("s3 %s: expected %q, got %v", key, want, got)
		}
	}
}
