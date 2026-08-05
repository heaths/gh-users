package github

import (
	"fmt"
	"io"
	"strings"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
)

const usersJQExpression = `[.data.repository[].nodes[]] | unique_by(.login) | sort_by(.login)`

type GraphQLClient interface {
	Do(query string, variables map[string]interface{}, response interface{}) error
}

type Client struct {
	gql GraphQLClient
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

type PageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type User struct {
	ID         string      `json:"id"`
	DatabaseID int         `json:"databaseId"`
	Login      string      `json:"login"`
	Name       string      `json:"name"`
	URL        string      `json:"url"`
	Email      string      `json:"email"`
	Status     *UserStatus `json:"status"`
}

type UserStatus struct {
	Message string `json:"message"`
}

func UsersJQExpression() string {
	return usersJQExpression
}

func New(log io.Writer) (*Client, error) {
	gql, err := gh.GQLClient(&api.ClientOptions{Log: log})
	if err != nil {
		return nil, err
	}

	return NewWithClient(gql), nil
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
		var response QueryResponse
		if err := c.gql.Do(query, vars, &response); err != nil {
			return nil, err
		}
		envelope.Data = response
		if envelope.Data.Repository == nil {
			envelope.Data.Repository = map[string]UserConnection{}
		}
		return envelope, nil
	}

	var endCursor string
	for {
		pagedVars := cloneVariables(vars)
		if endCursor != "" {
			pagedVars["endCursor"] = endCursor
		}

		var response QueryResponse
		if err := c.gql.Do(query, pagedVars, &response); err != nil {
			return nil, err
		}

		connection := response.Repository["alias_0"]
		aggregated := envelope.Data.Repository["alias_0"]
		aggregated.Nodes = append(aggregated.Nodes, connection.Nodes...)
		aggregated.PageInfo = connection.PageInfo
		envelope.Data.Repository["alias_0"] = aggregated

		if !connection.PageInfo.HasNextPage {
			break
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
	query.WriteString("  id\n")
	query.WriteString("  databaseId\n")
	query.WriteString("  login\n")
	query.WriteString("  name\n")
	query.WriteString("  url\n")
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
