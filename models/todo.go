package models

import "time"

type Todo struct {
	UserID      uint       `gorm:"primaryKey" json:"id"`
	ID          uint       `gorm:"primaryKey" json:"id"`
	Title       string     `gorm:"not null" json:"title" binding:"required"`
	Completed   bool       `gorm:"default:false" json:"completed"`
	Category    string     `json:"category" binding:"required"`
	Priority    string     `json:"priority" binding:"required"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}
