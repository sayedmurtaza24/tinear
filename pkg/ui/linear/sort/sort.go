package sort

import "github.com/sayedmurtaza24/tinear/linear/models"

type (
	SortBy    int
	SortOrder int
)

const (
	SortOrderDescending SortOrder = iota
	SortOrderAscending
)

const (
	SortByCreatedAt SortBy = iota
	SortByTitle
	SortByPriority
	SortByAssignee
	SortByState
)

type SortOption struct {
	sortBy    SortBy
	sortOrder SortOrder
}

func New(sortBy SortBy, sortOrder SortOrder) *SortOption {
	return &SortOption{
		sortBy:    sortBy,
		sortOrder: sortOrder,
	}
}

func (s *SortOption) ToIssueSortInput() []*models.IssueSortInput {
	var paginationSort models.PaginationSortOrder

	switch s.sortOrder {
	case SortOrderDescending:
		paginationSort = models.PaginationSortOrderDescending
	case SortOrderAscending:
		paginationSort = models.PaginationSortOrderAscending
	}

	var sortOptions models.IssueSortInput

	switch s.sortBy {
	case SortByCreatedAt:
		sortOptions = models.IssueSortInput{
			Priority: &models.PrioritySort{
				Order: &paginationSort,
			},
		}
	case SortByTitle:
		sortOptions = models.IssueSortInput{
			Title: &models.TitleSort{
				Order: &paginationSort,
			},
		}
	case SortByPriority:
		sortOptions = models.IssueSortInput{
			Priority: &models.PrioritySort{
				Order: &paginationSort,
			},
		}
	case SortByAssignee:
		sortOptions = models.IssueSortInput{
			Assignee: &models.AssigneeSort{
				Order: &paginationSort,
			},
		}
	case SortByState:
		sortOptions = models.IssueSortInput{
			WorkflowState: &models.WorkflowStateSort{
				Order: &paginationSort,
			},
		}
	}

	return []*models.IssueSortInput{&sortOptions}
}
