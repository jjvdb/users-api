package models

import "strings"

type LoginRequest struct {
	EmailOrUsername string  `json:"emailorusername"`
	Password        string  `json:"password"`
	Remember        bool    `json:"remember"`
	Device          *string `json:"device"`
	Location        *string `json:"location"`
}

func (req *LoginRequest) Trim() {
	req.EmailOrUsername = strings.TrimSpace(req.EmailOrUsername)
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewInternalError() ErrorResponse {
	return ErrorResponse{Error: "Something went wrong, try again later."}
}

func NewInvalidRequestBodyError() ErrorResponse {
	return ErrorResponse{Error: "Invalid request body."}
}

type SignupRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req *SignupRequest) Trim() {
	req.Name = strings.TrimSpace(req.Name)
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
}

type ParallelTranslationResponse struct {
	SourceTranslation    string   `json:"source_translation"`
	ParallelTranslations []string `json:"parallel_translations"`
}

type StatusType string

const (
	StatusComplete   StatusType = "complete"
	StatusPartial    StatusType = "partial"
	StatusNotStarted StatusType = "not_started"
)

// ReadBook represents a book reading progress
type ReadBook struct {
	Book         string     `json:"book"`
	Abbreviation string     `json:"abbreviation"`
	Status       StatusType `json:"status"`
}

type GenericMessage struct {
	Message string `json:"message"`
}

type BibleChapter struct {
	BookID       uint   `json:"book_id"`
	Book         string `json:"book"`
	Abbreviation string `json:"abbreviation"`
	Chapter      uint   `json:"chapter"`
}

type MarkChapterAsReadResponse struct {
	Book         string `json:"book"`
	Abbreviation string `json:"abbreviation"`
	Chapter      uint   `json:"chapter"`
	Message      string `json:"message"`
}

type MarkBookReadResponse struct {
	Message      string `json:"message"`
	Book         string `json:"book"`
	Abbreviation string `json:"abbreviation"`
	Count        int    `json:"count"`
}

type BookReadChaptersResponse struct {
	BookID       uint   `json:"book_id"`
	Book         string `json:"book"`
	Abbreviation string `json:"abbreviation"`
	ReadChapters []uint `json:"read_chapters"`
}