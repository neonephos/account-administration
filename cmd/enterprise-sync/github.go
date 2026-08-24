package main

import (
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const graphqlEndpoint = "https://api.github.com/graphql"

// client is a minimal GitHub GraphQL client for the enterprise APIs.
type client struct {
	token string
	http  *http.Client
}

func newClient(token string) *client {
	return &client{
		token: token,
		http:  &http.Client{Timeout: 30 * time.Second},
	}
}

type graphQLError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (e graphQLError) Error() string { return e.Message }

// do executes a GraphQL query/mutation and unmarshals `data` into out.
func (c *client) do(ctx context.Context, query string, vars map[string]any, out any) error {
	body, err := json.Marshal(map[string]any{"query": query, "variables": vars})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphqlEndpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("graphql http %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var envelope struct {
		Data   jsontext.Value `json:"data"`
		Errors []graphQLError `json:"errors"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if len(envelope.Errors) > 0 {
		msgs := make([]string, 0, len(envelope.Errors))
		for _, e := range envelope.Errors {
			msgs = append(msgs, e.Message)
		}
		return fmt.Errorf("graphql errors: %s", strings.Join(msgs, "; "))
	}
	if out != nil {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("decode data: %w", err)
		}
	}
	return nil
}

// enterpriseID resolves the node ID of an enterprise from its slug. Mutations
// require the node ID, not the slug.
func (c *client) enterpriseID(ctx context.Context, slug string) (string, error) {
	const q = `query($slug:String!){ enterprise(slug:$slug){ id } }`
	var out struct {
		Enterprise struct {
			ID string `json:"id"`
		} `json:"enterprise"`
	}
	if err := c.do(ctx, q, map[string]any{"slug": slug}, &out); err != nil {
		return "", err
	}
	if out.Enterprise.ID == "" {
		return "", fmt.Errorf("enterprise %q not found (or token lacks access)", slug)
	}
	return out.Enterprise.ID, nil
}

// liveState holds the current owners and members of the enterprise (lowercased
// login -> original login for reporting).
type liveState struct {
	owners  map[string]string
	members map[string]string
}

// fetchState lists current enterprise admins (owners) and members, paginating.
func (c *client) fetchState(ctx context.Context, slug string) (*liveState, error) {
	st := &liveState{owners: map[string]string{}, members: map[string]string{}}

	// Owners (admins) via ownerInfo.admins.
	const ownersQ = `query($slug:String!,$cursor:String){
      enterprise(slug:$slug){
        ownerInfo{
          admins(first:100, after:$cursor){
            nodes{ login }
            pageInfo{ hasNextPage endCursor }
          }
        }
      }
    }`
	cursor := (*string)(nil)
	for {
		var out struct {
			Enterprise struct {
				OwnerInfo struct {
					Admins struct {
						Nodes    []struct{ Login string `json:"login"` } `json:"nodes"`
						PageInfo struct {
							HasNextPage bool   `json:"hasNextPage"`
							EndCursor   string `json:"endCursor"`
						} `json:"pageInfo"`
					} `json:"admins"`
				} `json:"ownerInfo"`
			} `json:"enterprise"`
		}
		vars := map[string]any{"slug": slug}
		if cursor != nil {
			vars["cursor"] = *cursor
		}
		if err := c.do(ctx, ownersQ, vars, &out); err != nil {
			return nil, fmt.Errorf("list owners: %w", err)
		}
		for _, n := range out.Enterprise.OwnerInfo.Admins.Nodes {
			st.owners[strings.ToLower(n.Login)] = n.Login
		}
		pi := out.Enterprise.OwnerInfo.Admins.PageInfo
		if !pi.HasNextPage {
			break
		}
		c2 := pi.EndCursor
		cursor = &c2
	}

	// Members via enterprise.members (EnterpriseMemberEdge -> User).
	const membersQ = `query($slug:String!,$cursor:String){
      enterprise(slug:$slug){
        members(first:100, after:$cursor){
          nodes{ ... on User { login } }
          pageInfo{ hasNextPage endCursor }
        }
      }
    }`
	cursor = nil
	for {
		var out struct {
			Enterprise struct {
				Members struct {
					Nodes    []struct{ Login string `json:"login"` } `json:"nodes"`
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
				} `json:"members"`
			} `json:"enterprise"`
		}
		vars := map[string]any{"slug": slug}
		if cursor != nil {
			vars["cursor"] = *cursor
		}
		if err := c.do(ctx, membersQ, vars, &out); err != nil {
			return nil, fmt.Errorf("list members: %w", err)
		}
		for _, n := range out.Enterprise.Members.Nodes {
			if n.Login == "" {
				continue // non-User member node
			}
			lower := strings.ToLower(n.Login)
			// Owners also appear in members; keep them classified as owners only.
			if _, isOwner := st.owners[lower]; !isOwner {
				st.members[lower] = n.Login
			}
		}
		pi := out.Enterprise.Members.PageInfo
		if !pi.HasNextPage {
			break
		}
		c2 := pi.EndCursor
		cursor = &c2
	}

	return st, nil
}
