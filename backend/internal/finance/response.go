package finance

import "github.com/google/uuid"

type PaginationMeta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

type ExpenseResponse struct {
	ID            string         `json:"id"`
	CategoryID    *string        `json:"category_id,omitempty"`
	Category      string         `json:"category,omitempty"`
	CategoryName  string         `json:"category_name,omitempty"`
	Title         string         `json:"title"`
	Amount        int64          `json:"amount"`
	ExpenseDate   string         `json:"expense_date"`
	PaymentMethod string         `json:"payment_method,omitempty"`
	Note          string         `json:"note,omitempty"`
	CreatedBy     *ActorResponse `json:"created_by,omitempty"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

type ActorResponse struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

func NewExpenseResponse(expense Expense) ExpenseResponse {
	var categoryID *string
	if expense.CategoryID != nil {
		value := expense.CategoryID.String()
		categoryID = &value
	}

	var actor *ActorResponse
	if expense.CreatedBy != nil && *expense.CreatedBy != uuid.Nil {
		actor = &ActorResponse{
			ID:   expense.CreatedBy.String(),
			Name: expense.CreatedByName,
		}
	}

	return ExpenseResponse{
		ID:            expense.ID.String(),
		CategoryID:    categoryID,
		Category:      expense.CategorySlug,
		CategoryName:  expense.CategoryName,
		Title:         expense.Title,
		Amount:        expense.Amount,
		ExpenseDate:   expense.ExpenseDate.Format(dateFormat),
		PaymentMethod: expense.PaymentMethod,
		Note:          expense.Note,
		CreatedBy:     actor,
		CreatedAt:     expense.CreatedAt.Format(timeFormatRFC3339),
		UpdatedAt:     expense.UpdatedAt.Format(timeFormatRFC3339),
	}
}

const (
	dateFormat        = "2006-01-02"
	timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"
)
