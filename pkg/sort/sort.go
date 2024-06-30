package sort

//
// import (
// 	"slices"
// 	"strings"
// )
//
// type (
// 	SortBy    int
// 	SortOrder int
// )
//
// const (
// 	SortOrderDescending SortOrder = iota
// 	SortOrderAscending
// )
//
// const (
// 	SortSmart SortBy = iota
// 	SortByCreatedAt
// 	SortByUpdatedAt
// 	SortByTitle
// 	SortByPriority
// 	SortByAssignee
// 	SortByState
// )
//
// type SortOption struct {
// 	sortBy    SortBy
// 	sortOrder SortOrder
// }
//
// func New(sortBy SortBy, sortOrder SortOrder) *SortOption {
// 	return &SortOption{
// 		sortBy:    sortBy,
// 		sortOrder: sortOrder,
// 	}
// }
//
// func (s *SortOption) SortIssues(issues []issue.Issue) []issue.Issue {
// 	sortFunc := func(i, j issue.Issue) int {
// 		return 0
// 	}
//
// 	switch s.sortBy {
// 	case SortSmart:
// 		sortFunc = func(i, j issue.Issue) int {
// 			prioWeight := int(i.Priority - j.Priority)
// 			stateWeight := i.State.Position - j.State.Position
// 			isMeWeight := i.Assignee.SortWeight() - j.Assignee.SortWeight()
//
// 			return 16*isMeWeight + 8*prioWeight + 2*stateWeight
// 		}
// 	case SortByCreatedAt:
// 		sortFunc = func(i, j issue.Issue) int {
// 			return int(i.CreatedAt.Sub(j.CreatedAt))
// 		}
// 	case SortByUpdatedAt:
// 		sortFunc = func(i, j issue.Issue) int {
// 			return int(i.UpdatedAt.Compare(j.CreatedAt))
// 		}
// 	case SortByTitle:
// 		sortFunc = func(i, j issue.Issue) int {
// 			return strings.Compare(i.Title, j.Title)
// 		}
// 	case SortByPriority:
// 		sortFunc = func(i, j issue.Issue) int {
// 			return int(i.Priority - j.Priority)
// 		}
// 	case SortByAssignee:
// 		sortFunc = func(i, j issue.Issue) int {
// 			return strings.Compare(i.Assignee.DisplayName, j.Assignee.DisplayName)
// 		}
// 	case SortByState:
// 		sortFunc = func(i, j issue.Issue) int {
// 			return int(i.State.Position - j.State.Position)
// 		}
// 	}
//
// 	sorted := slices.Clone(issues)
//
// 	slices.SortStableFunc(sorted, func(i, j issue.Issue) int {
// 		if s.sortOrder == SortOrderDescending {
// 			return -sortFunc(i, j)
// 		}
// 		return sortFunc(i, j)
// 	})
//
// 	return issues
// }
