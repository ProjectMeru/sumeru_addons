package models

type LeadType string

const (
	LeadTypeLead         LeadType = "lead"
	LeadTypeOpportunity  LeadType = "opportunity"
)

type WonStatus string

const (
	WonStatusPending WonStatus = "pending"
	WonStatusWon     WonStatus = "won"
	WonStatusLost    WonStatus = "lost"
)

type LeadPriority string

const (
	LeadPriorityLow      LeadPriority = "0"
	LeadPriorityMedium   LeadPriority = "1"
	LeadPriorityHigh     LeadPriority = "2"
	LeadPriorityVeryHigh LeadPriority = "3"
)
