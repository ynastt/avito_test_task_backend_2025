package reviewers

import (
	"fmt"
	"math/rand/v2"

	"github.com/ynastt/avito_test_task_backend_2025/internal/domain"
)

func ChooseRandomReviewer(candidates []domain.User) (domain.User, error) {
	candidates_cnt := len(candidates)
	if candidates_cnt == 0 {
		return domain.User{}, fmt.Errorf("zero candidates availiable")
	}
	return candidates[rand.IntN(candidates_cnt)], nil
}

func ChooseRandomReviewers(candidates []domain.User, maxCount int) []domain.User {
	candidates_cnt := len(candidates)
	if candidates_cnt == 0 {
		return []domain.User{}
	}

	count := min(candidates_cnt, maxCount)

	result := make([]domain.User, 0, count)
	//  pseudo-random permutation of the integers in the half-open interval [0,candidates_cnt)
	indices := rand.Perm(candidates_cnt)

	for i := 0; i < count; i++ {
		result = append(result, candidates[indices[i]])
	}

	return result
}
