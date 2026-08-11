package github

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	ghterminal "github.com/heaths/gh-users/internal/terminal"
)

const usersJQExpression = `[.data.repository[].nodes[]] | unique_by(.login) | sort_by(.login)`

var userFields = []string{
	"login",
	"name",
	"email",
	"status",
}

type GraphQLClient interface {
	Do(query string, variables map[string]interface{}, response interface{}) error
}

type Client struct {
	gql        GraphQLClient
	newSpinner func() progressIndicator
}

type progressIndicator interface {
	Start()
	Stop()
}

type QueryEnvelope struct {
	Data QueryResponse `json:"data"`
}

type QueryResponse struct {
	Repository map[string]UserConnection `json:"repository"`
}

type UserConnection struct {
	Nodes    []User   `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

type graphqlQueryResponse struct {
	Repository map[string]graphqlUserConnection `json:"repository"`
}

type graphqlUserConnection struct {
	Nodes    []graphqlUser `json:"nodes"`
	PageInfo PageInfo      `json:"pageInfo"`
}

type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type User struct {
	Login  string `json:"login"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

type graphqlUser struct {
	Login  string             `json:"login"`
	Name   string             `json:"name"`
	Email  string             `json:"email"`
	Status *graphqlUserStatus `json:"status"`
}

type graphqlUserStatus struct {
	Message string `json:"message"`
}

func UserFields() []string {
	return slices.Clone(userFields)
}

func ParseUserFields(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return UserFields(), nil
	}

	fields := strings.Split(value, ",")
	for i, field := range fields {
		fields[i] = strings.TrimSpace(field)
	}

	if err := ValidateUserFields(fields); err != nil {
		return nil, err
	}

	return fields, nil
}

func ValidateUserFields(fields []string) error {
	for _, field := range fields {
		if field == "" {
			return fmt.Errorf("JSON fields cannot be empty")
		}
		if !slices.Contains(userFields, field) {
			return fmt.Errorf("unknown JSON field %q (available: %s)", field, strings.Join(userFields, ","))
		}
	}

	return nil
}

func (u User) ExportData(fields []string) (map[string]interface{}, error) {
	if err := ValidateUserFields(fields); err != nil {
		return nil, err
	}

	data := make(map[string]interface{}, len(fields))
	for _, field := range fields {
		switch field {
		case "login":
			data[field] = u.Login
		case "name":
			data[field] = u.Name
		case "email":
			data[field] = u.Email
		case "status":
			data[field] = u.Status
		}
	}

	return data, nil
}

func (u graphqlUser) User() User {
	user := User{
		Login: u.Login,
		Name:  u.Name,
		Email: u.Email,
	}
	if u.Status != nil {
		user.Status = u.Status.Message
	}

	return user
}

func (c graphqlUserConnection) UserConnection() UserConnection {
	connection := UserConnection{
		Nodes:    make([]User, 0, len(c.Nodes)),
		PageInfo: c.PageInfo,
	}
	for _, node := range c.Nodes {
		connection.Nodes = append(connection.Nodes, node.User())
	}

	return connection
}

func (r graphqlQueryResponse) QueryResponse() QueryResponse {
	response := QueryResponse{
		Repository: make(map[string]UserConnection, len(r.Repository)),
	}
	for alias, connection := range r.Repository {
		response.Repository[alias] = connection.UserConnection()
	}

	return response
}

func ExportUsers(users []User, fields []string) ([]map[string]interface{}, error) {
	if err := ValidateUserFields(fields); err != nil {
		return nil, err
	}

	data := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		exported, err := user.ExportData(fields)
		if err != nil {
			return nil, err
		}
		data = append(data, exported)
	}

	return data, nil
}

func UsersJQExpression() string {
	return usersJQExpression
}

func New(log io.Writer, spinnerOutput io.Writer) (*Client, error) {
	gql, err := api.NewGraphQLClient(api.ClientOptions{Log: log})
	if err != nil {
		return nil, err
	}

	return &Client{
		gql: gql,
		newSpinner: func() progressIndicator {
			return ghterminal.NewSpinner(spinnerOutput)
		},
	}, nil
}

func NewWithClient(gql GraphQLClient) *Client {
	return &Client{gql: gql}
}

func (c *Client) QueryUsers(owner, repo string, partials []string) (*QueryEnvelope, error) {
	query := BuildQuery(partials)
	vars := map[string]interface{}{
		"owner": owner,
		"repo":  repo,
	}
	for i, partial := range partials {
		vars[fmt.Sprintf("user_%d", i)] = partial
	}

	envelope := &QueryEnvelope{
		Data: QueryResponse{
			Repository: map[string]UserConnection{},
		},
	}

	if len(partials) > 0 {
		var response graphqlQueryResponse
		if err := c.gql.Do(query, vars, &response); err != nil {
			return nil, err
		}
		envelope.Data = response.QueryResponse()
		if envelope.Data.Repository == nil {
			envelope.Data.Repository = map[string]UserConnection{}
		}
		return envelope, nil
	}

	var endCursor string
	var indicator progressIndicator
	for {
		pagedVars := cloneVariables(vars)
		if endCursor != "" {
			pagedVars["endCursor"] = endCursor
		}

		var response graphqlQueryResponse
		if err := c.gql.Do(query, pagedVars, &response); err != nil {
			if indicator != nil {
				indicator.Stop()
			}
			return nil, err
		}

		connection := response.Repository["alias_0"].UserConnection()
		aggregated := envelope.Data.Repository["alias_0"]
		aggregated.Nodes = append(aggregated.Nodes, connection.Nodes...)
		aggregated.PageInfo = connection.PageInfo
		envelope.Data.Repository["alias_0"] = aggregated

		if !connection.PageInfo.HasNextPage {
			if indicator != nil {
				indicator.Stop()
			}
			break
		}

		if indicator == nil && c.newSpinner != nil {
			indicator = c.newSpinner()
			if indicator != nil {
				indicator.Start()
			}
		}

		endCursor = connection.PageInfo.EndCursor
	}

	return envelope, nil
}

func BuildQuery(partials []string) string {
	var query strings.Builder

	query.WriteString("query Users($owner: String!, $repo: String!")
	if len(partials) == 0 {
		query.WriteString(", $endCursor: String")
	}
	for i := range partials {
		fmt.Fprintf(&query, ", $user_%d: String!", i)
	}
	query.WriteString(") {\n")
	query.WriteString("  repository(owner: $owner, name: $repo) {\n")

	if len(partials) == 0 {
		query.WriteString("    alias_0: assignableUsers(first: 100, after: $endCursor) {\n")
		query.WriteString("      nodes {\n")
		query.WriteString("        ...UserFragment\n")
		query.WriteString("      }\n")
		query.WriteString("      pageInfo {\n")
		query.WriteString("        hasNextPage\n")
		query.WriteString("        endCursor\n")
		query.WriteString("      }\n")
		query.WriteString("    }\n")
	} else {
		for i := range partials {
			fmt.Fprintf(&query, "    alias_%d: assignableUsers(first: 100, query: $user_%d) {\n", i, i)
			query.WriteString("      nodes {\n")
			query.WriteString("        ...UserFragment\n")
			query.WriteString("      }\n")
			query.WriteString("    }\n")
		}
	}

	query.WriteString("  }\n")
	query.WriteString("}\n")
	query.WriteString("\n")
	query.WriteString("fragment UserFragment on User {\n")
	query.WriteString("  login\n")
	query.WriteString("  name\n")
	query.WriteString("  email\n")
	query.WriteString("  status {\n")
	query.WriteString("    message\n")
	query.WriteString("  }\n")
	query.WriteString("}\n")

	return query.String()
}

func cloneVariables(vars map[string]interface{}) map[string]interface{} {
	cloned := make(map[string]interface{}, len(vars))
	for key, value := range vars {
		cloned[key] = value
	}

	return cloned
}
