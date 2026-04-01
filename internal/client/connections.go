package client

import "fmt"

// Connection represents a Credible database connection.
type Connection struct {
	Name                    string               `json:"name,omitempty"`
	Type                    string               `json:"type,omitempty"`
	IncludeTables           []string             `json:"includeTables,omitempty"`
	ExcludeTables           []string             `json:"excludeTables,omitempty"`
	ExcludeAllTables        *bool                `json:"excludeAllTables,omitempty"`
	IndexingStatus          string               `json:"indexingStatus,omitempty"`
	PostgresConnection      *PostgresConnection  `json:"postgresConnection,omitempty"`
	BigqueryConnection      *BigqueryConnection  `json:"bigqueryConnection,omitempty"`
	SnowflakeConnection     *SnowflakeConnection `json:"snowflakeConnection,omitempty"`
	TrinoConnection         *TrinoConnection     `json:"trinoConnection,omitempty"`
	MysqlConnection         *MysqlConnection     `json:"mysqlConnection,omitempty"`
	DuckdbConnection        *DuckdbConnection    `json:"duckdbConnection,omitempty"`
	MotherduckConnection    *MotherduckConnection `json:"motherduckConnection,omitempty"`
}

type PostgresConnection struct {
	Host             string `json:"host,omitempty"`
	Port             *int   `json:"port,omitempty"`
	DatabaseName     string `json:"databaseName,omitempty"`
	UserName         string `json:"userName,omitempty"`
	Password         string `json:"password,omitempty"`
	ConnectionString string `json:"connectionString,omitempty"`
}

type BigqueryConnection struct {
	DefaultProjectId         string `json:"defaultProjectId,omitempty"`
	BillingProjectId         string `json:"billingProjectId,omitempty"`
	Location                 string `json:"location,omitempty"`
	ServiceAccountKeyJson    string `json:"serviceAccountKeyJson,omitempty"`
	MaximumBytesBilled       string `json:"maximumBytesBilled,omitempty"`
	QueryTimeoutMilliseconds string `json:"queryTimeoutMilliseconds,omitempty"`
}

type SnowflakeConnection struct {
	Account                      string `json:"account,omitempty"`
	Username                     string `json:"username,omitempty"`
	Password                     string `json:"password,omitempty"`
	PrivateKey                   string `json:"privateKey,omitempty"`
	PrivateKeyPass               string `json:"privateKeyPass,omitempty"`
	Warehouse                    string `json:"warehouse,omitempty"`
	Database                     string `json:"database,omitempty"`
	Schema                       string `json:"schema,omitempty"`
	Role                         string `json:"role,omitempty"`
	ResponseTimeoutMilliseconds  *int   `json:"responseTimeoutMilliseconds,omitempty"`
}

type TrinoConnection struct {
	Server  string `json:"server,omitempty"`
	Port    *int   `json:"port,omitempty"`
	Catalog string `json:"catalog,omitempty"`
	Schema  string `json:"schema,omitempty"`
	User    string `json:"user,omitempty"`
}

type MysqlConnection struct {
	Host     string `json:"host,omitempty"`
	Port     *int   `json:"port,omitempty"`
	Database string `json:"database,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

type DuckdbConnection struct {
	URL     string `json:"url,omitempty"`
	MdToken string `json:"md_token,omitempty"`
}

type MotherduckConnection struct {
	URL     string `json:"url,omitempty"`
	MdToken string `json:"md_token,omitempty"`
}

func (c *Client) ListConnections(org, project string) ([]Connection, error) {
	var result []Connection
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/projects/%s/connections", org, project), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing connections: %w", err)
	}
	return result, nil
}

func (c *Client) CreateConnection(org, project string, conn *Connection) (*Connection, error) {
	var result Connection
	err := c.doJSON("POST", fmt.Sprintf("/organizations/%s/projects/%s/connections", org, project), conn, &result)
	if err != nil {
		return nil, fmt.Errorf("creating connection: %w", err)
	}
	return &result, nil
}

func (c *Client) GetConnection(org, project, name string) (*Connection, error) {
	var result Connection
	err := c.doJSON("GET", fmt.Sprintf("/organizations/%s/projects/%s/connections/%s", org, project, name), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting connection %q: %w", name, err)
	}
	return &result, nil
}

func (c *Client) UpdateConnection(org, project, name string, conn *Connection) (*Connection, error) {
	var result Connection
	err := c.doJSON("PATCH", fmt.Sprintf("/organizations/%s/projects/%s/connections/%s", org, project, name), conn, &result)
	if err != nil {
		return nil, fmt.Errorf("updating connection %q: %w", name, err)
	}
	return &result, nil
}

func (c *Client) DeleteConnection(org, project, name string) error {
	err := c.doJSON("DELETE", fmt.Sprintf("/organizations/%s/projects/%s/connections/%s", org, project, name), nil, nil)
	if err != nil {
		return fmt.Errorf("deleting connection %q: %w", name, err)
	}
	return nil
}
