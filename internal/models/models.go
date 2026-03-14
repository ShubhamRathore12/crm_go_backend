package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Lead represents a sales lead
type Lead struct {
	ID             uuid.UUID      `json:"id"`
	Name           string         `json:"name"`
	Email          string         `json:"email"`
	Phone          string         `json:"phone"`
	Company        string         `json:"company"`
	Source         string         `json:"source"`
	Status         string         `json:"status"`
	AssignedTo     *uuid.UUID     `json:"assigned_to,omitempty"`
	TeamID         *uuid.UUID     `json:"team_id,omitempty"`
	Score          *int           `json:"score,omitempty"`
	CustomFields   map[string]any `json:"custom_fields,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	ArchivedAt     *time.Time     `json:"archived_at,omitempty"`
}

// Contact represents a contact
type Contact struct {
	ID           uuid.UUID      `json:"id"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Phone        string         `json:"phone"`
	Company      string         `json:"company"`
	LeadID       *uuid.UUID     `json:"lead_id,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	ArchivedAt   *time.Time     `json:"archived_at,omitempty"`
}

// Interaction represents an interaction with a lead/contact
type Interaction struct {
	ID          uuid.UUID  `json:"id"`
	ContactID   uuid.UUID  `json:"contact_id"`
	LeadID      uuid.UUID  `json:"lead_id"`
	Type        string     `json:"type"` // call, email, sms, whatsapp, note
	Direction   string     `json:"direction"` // inbound, outbound, internal
	Content     string     `json:"content"`
	Metadata    any        `json:"metadata,omitempty"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Workflow represents an automated workflow
type Workflow struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Trigger     any        `json:"trigger"`
	Actions     []any      `json:"actions"`
	Active      bool       `json:"active"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Integration represents an external integration
type Integration struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Config     any        `json:"config"`
	Active     bool       `json:"active"`
	CreatedBy  uuid.UUID  `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Team represents a team
type Team struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Members     []uuid.UUID `json:"members"`
	LeaderID    uuid.UUID  `json:"leader_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Opportunity represents a sales opportunity
type Opportunity struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	ContactID   uuid.UUID      `json:"contact_id"`
	Value       float64        `json:"value"`
	Stage       string         `json:"stage"`
	Probability int            `json:"probability"`
	ExpectedClose *time.Time   `json:"expected_close,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// AIScore represents an AI-generated score
type AIScore struct {
	ID          uuid.UUID  `json:"id"`
	EntityID    uuid.UUID  `json:"entity_id"`
	EntityType  string     `json:"entity_type"` // lead, contact, opportunity
	Score       int        `json:"score"`
	Factors     any        `json:"factors,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
