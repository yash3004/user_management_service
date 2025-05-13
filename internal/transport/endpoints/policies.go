package endpoints

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/policies"
)

// Policy represents a policy in the response
type Policy struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Effect      string    `json:"effect"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreatePolicyRequest represents the create policy request
type CreatePolicyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Effect      string `json:"effect"`
}

// CreatePolicyResponse represents the create policy response
type CreatePolicyResponse struct {
	Policy Policy `json:"policy"`
}

// GetPolicyRequest represents the get policy request
type GetPolicyRequest struct {
	ID string `json:"id"`
}

// GetPolicyResponse represents the get policy response
type GetPolicyResponse struct {
	Policy Policy `json:"policy"`
}

// ListPoliciesRequest represents the list policies request
type ListPoliciesRequest struct {
	// Add pagination parameters if needed
}

// ListPoliciesResponse represents the list policies response
type ListPoliciesResponse struct {
	Policies []Policy `json:"policies"`
}

// UpdatePolicyRequest represents the update policy request
type UpdatePolicyRequest struct {
	ID          string `json:"-"` // From URL path
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Effect      string `json:"effect"`
}

// UpdatePolicyResponse represents the update policy response
type UpdatePolicyResponse struct {
	Policy Policy `json:"policy"`
}

// DeletePolicyRequest represents the delete policy request
type DeletePolicyRequest struct {
	ID string `json:"id"`
}

// DeletePolicyResponse represents the delete policy response
type DeletePolicyResponse struct {
	Success bool `json:"success"`
}

// PoliciesEndpoint handles policy-related endpoints
type PoliciesEndpoint struct {
	PolicyManager policies.PolicyManager
}

// NewPoliciesEndpoint creates a new policies endpoint
func NewPoliciesEndpoint(manager policies.PolicyManager) *PoliciesEndpoint {
	return &PoliciesEndpoint{
		PolicyManager: manager,
	}
}

// CreatePolicy creates a new policy
func (e *PoliciesEndpoint) CreatePolicy(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(CreatePolicyRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Delegate to the policy manager
	policy, err := e.PolicyManager.CreatePolicy(ctx, req.Name, req.Description, req.Resource, req.Action, req.Effect)
	if err != nil {
		return nil, err
	}

	return CreatePolicyResponse{
		Policy: Policy{
			ID:          policy.ID.String(),
			Name:        policy.Name,
			Description: policy.Description,
			Resource:    policy.Resource,
			Action:      policy.Action,
			Effect:      policy.Effect,
			CreatedAt:   policy.CreatedAt,
			UpdatedAt:   policy.UpdatedAt,
		},
	}, nil
}

// GetPolicy gets a policy by ID
func (e *PoliciesEndpoint) GetPolicy(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(GetPolicyRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUID
	policyID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid policy ID format")
	}

	// Delegate to the policy manager
	policy, err := e.PolicyManager.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, err
	}

	return GetPolicyResponse{
		Policy: Policy{
			ID:          policy.ID.String(),
			Name:        policy.Name,
			Description: policy.Description,
			Resource:    policy.Resource,
			Action:      policy.Action,
			Effect:      policy.Effect,
			CreatedAt:   policy.CreatedAt,
			UpdatedAt:   policy.UpdatedAt,
		},
	}, nil
}

// ListPolicies lists all policies
func (e *PoliciesEndpoint) ListPolicies(ctx context.Context, request interface{}) (interface{}, error) {
	// Delegate to the policy manager
	policiesList, err := e.PolicyManager.ListPolicies(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	policies := make([]Policy, len(policiesList))
	for i, p := range policiesList {
		policies[i] = Policy{
			ID:          p.ID.String(),
			Name:        p.Name,
			Description: p.Description,
			Resource:    p.Resource,
			Action:      p.Action,
			Effect:      p.Effect,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
	}

	return ListPoliciesResponse{
		Policies: policies,
	}, nil
}

// UpdatePolicy updates a policy
func (e *PoliciesEndpoint) UpdatePolicy(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(UpdatePolicyRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUID
	policyID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid policy ID format")
	}

	// Delegate to the policy manager
	policy, err := e.PolicyManager.UpdatePolicy(ctx, policyID, req.Name, req.Description, req.Resource, req.Action, req.Effect)
	if err != nil {
		return nil, err
	}

	return UpdatePolicyResponse{
		Policy: Policy{
			ID:          policy.ID.String(),
			Name:        policy.Name,
			Description: policy.Description,
			Resource:    policy.Resource,
			Action:      policy.Action,
			Effect:      policy.Effect,
			CreatedAt:   policy.CreatedAt,
			UpdatedAt:   policy.UpdatedAt,
		},
	}, nil
}

// DeletePolicy deletes a policy
func (e *PoliciesEndpoint) DeletePolicy(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(DeletePolicyRequest)
	if !ok {
		return nil, errors.New("invalid request format")
	}

	// Parse UUID
	policyID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, errors.New("invalid policy ID format")
	}

	// Delegate to the policy manager
	err = e.PolicyManager.DeletePolicy(ctx, policyID)
	if err != nil {
		return nil, err
	}

	return DeletePolicyResponse{
		Success: true,
	}, nil
}