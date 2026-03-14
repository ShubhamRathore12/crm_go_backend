package ai

import (
	"crm-backend/internal/db"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Scorer handles AI scoring operations
type Scorer struct {
	db *db.DatabaseManager
}

// Score represents an AI score
type Score struct {
	ID         uuid.UUID      `json:"id"`
	EntityID   uuid.UUID      `json:"entity_id"`
	EntityType string         `json:"entity_type"`
	Score      int            `json:"score"`
	Factors    map[string]any `json:"factors,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// ScoreRequest represents a scoring request
type ScoreRequest struct {
	EntityID   string                 `json:"entity_id" binding:"required"`
	EntityType string                 `json:"entity_type" binding:"required"`
	Data       map[string]interface{} `json:"data"`
}

// NewScorer creates a new AI scorer
func NewScorer(db *db.DatabaseManager) *Scorer {
	return &Scorer{
		db: db,
	}
}

// CalculateScore calculates a score for an entity
func (s *Scorer) CalculateScore(entityID uuid.UUID, entityType string, data map[string]interface{}) (*Score, error) {
	log.Printf("Calculating %s score for %s", entityType, entityID.String())

	// TODO: Implement actual AI/ML scoring logic
	// For now, return a placeholder score
	score := &Score{
		ID:         uuid.New(),
		EntityID:   entityID,
		EntityType: entityType,
		Score:      75, // Placeholder
		Factors: map[string]any{
			"engagement": 70,
			"demographics": 80,
			"behavior": 75,
		},
		CreatedAt: time.Now(),
	}

	// Save to database
	if err := s.saveScore(score); err != nil {
		return nil, err
	}

	return score, nil
}

// saveScore saves a score to the database
func (s *Scorer) saveScore(score *Score) error {
	factorsJSON, _ := json.Marshal(score.Factors)
	
	_, err := s.db.WritePool().Exec(
		`INSERT INTO ai_scores (id, entity_id, entity_type, score, factors, created_at) 
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		score.ID, score.EntityID, score.EntityType, score.Score, factorsJSON, score.CreatedAt,
	)
	return err
}

// GetScoresByEntity retrieves scores for an entity
func (s *Scorer) GetScoresByEntity(entityID uuid.UUID, entityType string) ([]Score, error) {
	rows, err := s.db.ReadPool().Query(
		`SELECT id, entity_id, entity_type, score, factors, created_at 
		 FROM ai_scores WHERE entity_id = $1 AND entity_type = $2 
		 ORDER BY created_at DESC LIMIT 10`,
		entityID, entityType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []Score
	for rows.Next() {
		var score Score
		var factorsJSON []byte
		err := rows.Scan(&score.ID, &score.EntityID, &score.EntityType, &score.Score, &factorsJSON, &score.CreatedAt)
		if err != nil {
			return nil, err
		}
		
		if factorsJSON != nil {
			json.Unmarshal(factorsJSON, &score.Factors)
		}
		
		scores = append(scores, score)
	}

	return scores, nil
}

// ScoreHandler handles POST /ai/score requests
func ScoreHandler(c *gin.Context) {
	var req ScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get scorer from context
	scorer, ok := c.Get("ai_scorer")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI scorer not available"})
		return
	}

	engine := scorer.(*Scorer)

	entityID, err := uuid.Parse(req.EntityID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	score, err := engine.CalculateScore(entityID, req.EntityType, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate score"})
		return
	}

	c.JSON(http.StatusOK, score)
}

// GetScoresHandler handles GET /ai/scores requests
func GetScoresHandler(c *gin.Context) {
	entityIDStr := c.Query("entity_id")
	entityType := c.Query("entity_type")

	if entityIDStr == "" || entityType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing entity_id or entity_type"})
		return
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity_id"})
		return
	}

	// Get scorer from context
	scorer, ok := c.Get("ai_scorer")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI scorer not available"})
		return
	}

	engine := scorer.(*Scorer)

	scores, err := engine.GetScoresByEntity(entityID, entityType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve scores"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"scores": scores})
}
