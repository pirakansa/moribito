package webhook

import (
	"github.com/pirakansa/moribito/internal/issue"
	"github.com/pirakansa/moribito/internal/review"
)

func resolveRepoServices(resolver RepoServiceResolver, repoFullName string) (review.Reviewer, review.PRCommenter, *issue.Service, bool) {
	if resolver == nil {
		return nil, nil, nil, false
	}
	return resolver(repoFullName)
}
