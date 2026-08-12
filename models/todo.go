package models

import "time"

type Todo struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserTaskID  uint       `json:"user_task_id"`
	UserID      uint       `json:"user_id"`
	User        User       `gorm:"foreignKey:UserID" json:"-"` // json:"-" بتمنع ظهور بيانات اليوزر بالكامل جوه التاسك
	Title       string     `gorm:"not null" json:"title" binding:"required"`
	Completed   bool       `gorm:"default:false" json:"completed"`
	Category    string     `json:"category" binding:"required"`
	Priority    string     `json:"priority" binding:"required"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}
