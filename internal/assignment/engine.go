package assignment

import (
	"crm-backend/internal/db"
	"log"

	"github.com/google/uuid"
)

// Engine handles lead/contact assignment
type Engine struct {
	db *db.DatabaseManager
}

// AssignmentStrategy defines how to assign entities
type AssignmentStrategy string

const (
	RoundRobin    AssignmentStrategy = "round_robin"
	BasedOnScore  AssignmentStrategy = "based_on_score"
	BasedOnLoad   AssignmentStrategy = "based_on_load"
	Manual        AssignmentStrategy = "manual"
)

// NewEngine creates a new assignment engine
func NewEngine(db *db.DatabaseManager) *Engine {
	return &Engine{
		db: db,
	}
}

// AssignLead assigns a lead to a user or team
func (e *Engine) AssignLead(leadID uuid.UUID, strategy AssignmentStrategy) error {
	log.Printf("Assigning lead %s using strategy %s", leadID.String(), strategy)

	switch strategy {
	case RoundRobin:
		return e.assignRoundRobin(leadID)
	case BasedOnScore:
		return e.assignBasedOnScore(leadID)
	case BasedOnLoad:
		return e.assignBasedOnLoad(leadID)
	default:
		log.Printf("Unknown strategy: %s", strategy)
		return nil
	}
}

// assignRoundRobin assigns to the next user in rotation
func (e *Engine) assignRoundRobin(leadID uuid.UUID) error {
	// TODO: Implement round-robin assignment
	log.Println("Round-robin assignment")
	return nil
}

// assignBasedOnScore assigns based on lead score
func (e *Engine) assignBasedOnScore(leadID uuid.UUID) error {
	// TODO: Implement score-based assignment
	log.Println("Score-based assignment")
	return nil
}

// assignBasedOnLoad assigns based on user workload
func (e *Engine) assignBasedOnLoad(leadID uuid.UUID) error {
	// TODO: Implement load-based assignment
	log.Println("Load-based assignment")
	return nil
}

// AssignToUser assigns a lead to a specific user
func (e *Engine) AssignToUser(leadID, userID uuid.UUID) error {
	_, err := e.db.WritePool().Exec(
		"UPDATE leads SET assigned_to = $1, updated_at = NOW() WHERE id = $2",
		userID, leadID,
	)
	return err
}

// AssignToTeam assigns a lead to a team
func (e *Engine) AssignToTeam(leadID, teamID uuid.UUID) error {
	_, err := e.db.WritePool().Exec(
		"UPDATE leads SET team_id = $1, updated_at = NOW() WHERE id = $2",
		teamID, leadID,
	)
	return err
}
