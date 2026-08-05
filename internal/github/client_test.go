package github

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeGQLClient struct {
	do func(query string, vars map[string]interface{}, response interface{}) error
}

func (f *fakeGQLClient) Do(query string, vars map[string]interface{}, response interface{}) error {
	return f.do(query, vars, response)
}

func TestBuildQuery_MultiplePartialsUsesAliasesAndFragment(t *testing.T) {
	query := BuildQuery([]string{"heath", "octo"})

	require.Contains(t, query, "query Users($owner: String!, $repo: String!, $user_0: String!, $user_1: String!)")
	require.Contains(t, query, "alias_0: assignableUsers(first: 100, query: $user_0)")
	require.Contains(t, query, "alias_1: assignableUsers(first: 100, query: $user_1)")
	require.Contains(t, query, "fragment UserFragment on User")
	require.NotContains(t, query, "\n  id\n")
	require.NotContains(t, query, "databaseId")
	require.NotContains(t, query, "url")
}

func TestBuildQuery_NoPartialsUsesPagination(t *testing.T) {
	query := BuildQuery(nil)

	require.Contains(t, query, "$endCursor: String")
	require.Contains(t, query, "alias_0: assignableUsers(first: 100, after: $endCursor)")
	require.Contains(t, query, "pageInfo")
}

func TestQueryUsers_UsesBatchedAliases(t *testing.T) {
	client := NewWithClient(&fakeGQLClient{
		do: func(query string, vars map[string]interface{}, response interface{}) error {
			require.Equal(t, "heaths", vars["owner"])
			require.Equal(t, "gh-users", vars["repo"])
			require.Equal(t, "heath", vars["user_0"])
			require.Equal(t, "octo", vars["user_1"])

			resp := response.(*graphqlQueryResponse)
			resp.Repository = map[string]graphqlUserConnection{
				"alias_0": {Nodes: []graphqlUser{{Login: "heaths", Status: &graphqlUserStatus{Message: "available"}}}},
				"alias_1": {Nodes: []graphqlUser{{Login: "octocat"}}},
			}
			return nil
		},
	})

	response, err := client.QueryUsers("heaths", "gh-users", []string{"heath", "octo"})
	require.NoError(t, err)
	require.Len(t, response.Data.Repository, 2)
	require.Equal(t, "heaths", response.Data.Repository["alias_0"].Nodes[0].Login)
	require.Equal(t, "available", response.Data.Repository["alias_0"].Nodes[0].Status)
	require.Equal(t, "octocat", response.Data.Repository["alias_1"].Nodes[0].Login)
}

func TestQueryUsers_PaginatesUnfilteredResults(t *testing.T) {
	calls := 0
	client := NewWithClient(&fakeGQLClient{
		do: func(query string, vars map[string]interface{}, response interface{}) error {
			calls++
			resp := response.(*graphqlQueryResponse)
			switch calls {
			case 1:
				_, hasCursor := vars["endCursor"]
				require.False(t, hasCursor)
				resp.Repository = map[string]graphqlUserConnection{
					"alias_0": {
						Nodes: []graphqlUser{{Login: "alpha"}},
						PageInfo: PageInfo{
							HasNextPage: true,
							EndCursor:   "cursor-1",
						},
					},
				}
			case 2:
				require.Equal(t, "cursor-1", vars["endCursor"])
				resp.Repository = map[string]graphqlUserConnection{
					"alias_0": {
						Nodes: []graphqlUser{{Login: "beta"}},
						PageInfo: PageInfo{
							HasNextPage: false,
						},
					},
				}
			default:
				return errors.New("unexpected page")
			}

			return nil
		},
	})

	response, err := client.QueryUsers("heaths", "gh-users", nil)
	require.NoError(t, err)
	require.Equal(t, 2, calls)
	require.Equal(t, []User{{Login: "alpha"}, {Login: "beta"}}, response.Data.Repository["alias_0"].Nodes)
}

func TestParseUserFields(t *testing.T) {
	fields, err := ParseUserFields("login, status")
	require.NoError(t, err)
	require.Equal(t, []string{"login", "status"}, fields)
}

func TestParseUserFields_RejectsUnknownField(t *testing.T) {
	_, err := ParseUserFields("login,url")
	require.EqualError(t, err, `unknown JSON field "url" (available: login,name,email,status)`)
}
