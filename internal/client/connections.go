package client

import "fmt"

// Connection represents a Credible database connection.
//
// Exactly one of the per-engine sub-objects is populated, selected by Type. The
// wrapper key for MotherDuck is "motherduckConnection" (lowercase d) even though
// its schema is MotherDuckConnection -- that asymmetry is in the API spec, so the
// tag must not be "corrected" to match the Go type name.
type Connection struct {
	Name                 string                `json:"name,omitempty"`
	Type                 string                `json:"type,omitempty"`
	IncludeTables        []string              `json:"includeTables,omitempty"`
	ExcludeTables        []string              `json:"excludeTables,omitempty"`
	ExcludeAllTables     *bool                 `json:"excludeAllTables,omitempty"`
	IndexingStatus       string                `json:"indexingStatus,omitempty"`
	Proxy                *ConnectionProxy      `json:"proxy,omitempty"`
	PostgresConnection   *PostgresConnection   `json:"postgresConnection,omitempty"`
	BigqueryConnection   *BigqueryConnection   `json:"bigqueryConnection,omitempty"`
	SnowflakeConnection  *SnowflakeConnection  `json:"snowflakeConnection,omitempty"`
	TrinoConnection      *TrinoConnection      `json:"trinoConnection,omitempty"`
	DatabricksConnection *DatabricksConnection `json:"databricksConnection,omitempty"`
	MysqlConnection      *MysqlConnection      `json:"mysqlConnection,omitempty"`
	DuckdbConnection     *DuckdbConnection     `json:"duckdbConnection,omitempty"`
	MotherDuckConnection *MotherDuckConnection `json:"motherduckConnection,omitempty"`
	DucklakeConnection   *DucklakeConnection   `json:"ducklakeConnection,omitempty"`
	PublisherConnection  *PublisherConnection  `json:"publisherConnection,omitempty"`
}

// ConnectionProxy reaches a database that is not directly routable (e.g. behind a
// bastion). Modeled as a discriminated union on Type so further mechanisms can be
// added; SSH local port-forwarding is the only one the API accepts today.
type ConnectionProxy struct {
	Type string          `json:"type,omitempty"`
	Ssh  *SshProxyConfig `json:"ssh,omitempty"`
}

// SshProxyConfig authenticates to the bastion by public key only.
//
// PrivateKey and PrivateKeyPass are write-only: reads never return them, and
// sending an empty value on update means "keep the stored one" rather than
// "clear it".
type SshProxyConfig struct {
	Host           string `json:"host,omitempty"`
	Port           *int   `json:"port,omitempty"`
	Username       string `json:"username,omitempty"`
	PrivateKey     string `json:"privateKey,omitempty"`
	PrivateKeyPass string `json:"privateKeyPass,omitempty"`
	HostKey        string `json:"hostKey,omitempty"`
}

type PostgresConnection struct {
	Host             string `json:"host,omitempty"`
	Port             *int   `json:"port,omitempty"`
	DatabaseName     string `json:"databaseName,omitempty"`
	UserName         string `json:"userName,omitempty"`
	Password         string `json:"password,omitempty"`
	ConnectionString string `json:"connectionString,omitempty"`
	// Sslmode is only valid on a proxied connection: a direct connection uses the
	// deployment's PGSSLMODE and the API rejects the field outright.
	Sslmode string `json:"sslmode,omitempty"`
}

type BigqueryConnection struct {
	DefaultProjectId         string `json:"defaultProjectId,omitempty"`
	BillingProjectId         string `json:"billingProjectId,omitempty"`
	Location                 string `json:"location,omitempty"`
	ServiceAccountKeyJson    string `json:"serviceAccountKeyJson,omitempty"`
	MaximumBytesBilled       string `json:"maximumBytesBilled,omitempty"`
	QueryTimeoutMilliseconds string `json:"queryTimeoutMilliseconds,omitempty"`
	// ImpersonateServiceAccount is mutually exclusive with ServiceAccountKeyJson.
	ImpersonateServiceAccount string `json:"impersonateServiceAccount,omitempty"`
}

type SnowflakeConnection struct {
	Account                     string `json:"account,omitempty"`
	Username                    string `json:"username,omitempty"`
	Password                    string `json:"password,omitempty"`
	PrivateKey                  string `json:"privateKey,omitempty"`
	PrivateKeyPass              string `json:"privateKeyPass,omitempty"`
	Warehouse                   string `json:"warehouse,omitempty"`
	Database                    string `json:"database,omitempty"`
	Schema                      string `json:"schema,omitempty"`
	Role                        string `json:"role,omitempty"`
	ResponseTimeoutMilliseconds *int   `json:"responseTimeoutMilliseconds,omitempty"`
}

type TrinoConnection struct {
	Server   string `json:"server,omitempty"`
	Port     *int   `json:"port,omitempty"`
	Catalog  string `json:"catalog,omitempty"`
	Schema   string `json:"schema,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	PeakaKey string `json:"peakaKey,omitempty"`
}

type DatabricksConnection struct {
	Host              string `json:"host,omitempty"`
	Path              string `json:"path,omitempty"`
	Token             string `json:"token,omitempty"`
	OauthClientId     string `json:"oauthClientId,omitempty"`
	OauthClientSecret string `json:"oauthClientSecret,omitempty"`
	DefaultCatalog    string `json:"defaultCatalog,omitempty"`
	DefaultSchema     string `json:"defaultSchema,omitempty"`
	SetupSQL          string `json:"setupSQL,omitempty"`
}

type MysqlConnection struct {
	Host     string `json:"host,omitempty"`
	Port     *int   `json:"port,omitempty"`
	Database string `json:"database,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// DuckdbConnection exposes data-source intent only. Database files, working
// directories, filesystem/network policy, extension loading and resource knobs are
// owned by Publisher so an environment config cannot widen deployment policy.
type DuckdbConnection struct {
	AttachedDatabases []AttachedDatabase `json:"attachedDatabases,omitempty"`
}

// AttachedDatabase is a warehouse attached to a DuckDB connection. Type selects
// which of the sub-objects applies.
type AttachedDatabase struct {
	Name                string               `json:"name,omitempty"`
	Type                string               `json:"type,omitempty"`
	BigqueryConnection  *BigqueryConnection  `json:"bigqueryConnection,omitempty"`
	SnowflakeConnection *SnowflakeConnection `json:"snowflakeConnection,omitempty"`
	PostgresConnection  *PostgresConnection  `json:"postgresConnection,omitempty"`
	GcsConnection       *GCSConnection       `json:"gcsConnection,omitempty"`
	S3Connection        *S3Connection        `json:"s3Connection,omitempty"`
	AzureConnection     *AzureConnection     `json:"azureConnection,omitempty"`
}

type MotherDuckConnection struct {
	AccessToken string `json:"accessToken,omitempty"`
	Database    string `json:"database,omitempty"`
}

type PublisherConnection struct {
	ConnectionUri string `json:"connectionUri,omitempty"`
	AccessToken   string `json:"accessToken,omitempty"`
}

type DucklakeConnection struct {
	Storage *DucklakeStorage `json:"storage,omitempty"`
	Catalog *DucklakeCatalog `json:"catalog,omitempty"`
}

type DucklakeStorage struct {
	BucketUrl     string         `json:"bucketUrl,omitempty"`
	S3Connection  *S3Connection  `json:"s3Connection,omitempty"`
	GcsConnection *GCSConnection `json:"gcsConnection,omitempty"`
}

type DucklakeCatalog struct {
	PostgresConnection *PostgresConnection `json:"postgresConnection,omitempty"`
	// MetadataSchema separates several DuckLake catalogs sharing one catalog
	// database. It is an organizational boundary, not an access-control one.
	MetadataSchema string `json:"metadataSchema,omitempty"`
}

type GCSConnection struct {
	KeyId  string `json:"keyId,omitempty"`
	Secret string `json:"secret,omitempty"`
}

type S3Connection struct {
	// Provider "config" requires AccessKeyId/SecretAccessKey; "credential_chain"
	// resolves them from the host and rejects a supplied key.
	Provider        string `json:"provider,omitempty"`
	Chain           string `json:"chain,omitempty"`
	AccessKeyId     string `json:"accessKeyId,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	Region          string `json:"region,omitempty"`
	Endpoint        string `json:"endpoint,omitempty"`
	SessionToken    string `json:"sessionToken,omitempty"`
}

type AzureConnection struct {
	AuthType     string `json:"authType,omitempty"`
	SasUrl       string `json:"sasUrl,omitempty"`
	TenantId     string `json:"tenantId,omitempty"`
	ClientId     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	AccountName  string `json:"accountName,omitempty"`
	FileUrl      string `json:"fileUrl,omitempty"`
}

func (c *Client) ListConnections(org, environment string) ([]Connection, error) {
	var result []Connection
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/environments/%s/connections", org, environment), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing connections: %w", err)
	}
	return result, nil
}

func (c *Client) CreateConnection(org, environment string, conn *Connection) (*Connection, error) {
	var result Connection
	err := c.doJSON("POST", fmt.Sprintf("/organizations/%s/environments/%s/connections", org, environment), conn, &result)
	if err != nil {
		return nil, fmt.Errorf("creating connection: %w", err)
	}
	return &result, nil
}

func (c *Client) GetConnection(org, environment, name string) (*Connection, error) {
	var result Connection
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/environments/%s/connections/%s", org, environment, name), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting connection %q: %w", name, err)
	}
	return &result, nil
}

func (c *Client) UpdateConnection(org, environment, name string, conn *Connection) (*Connection, error) {
	var result Connection
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s/environments/%s/connections/%s", org, environment, name), conn, &result)
	if err != nil {
		return nil, fmt.Errorf("updating connection %q: %w", name, err)
	}
	return &result, nil
}

func (c *Client) DeleteConnection(org, environment, name string) error {
	err := c.doJSON("DELETE", fmt.Sprintf("/organizations/%s/environments/%s/connections/%s", org, environment, name), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting connection %q: %w", name, err)
	}
	return nil
}
