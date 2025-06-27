package postgres

import "ndc/ai_bot/internal/repo"

// UseCase -.
type UseCase struct {
	Repo      repo.ProductRepo
	RepoOrder repo.OrderRepo
	AuthRepo  repo.AuthRepo
	Chat      repo.ChatRepo
	Business  repo.BusinessRepo
}

// New -.
func New(us UseCase) *UseCase {
	return &UseCase{
		Repo:      us.Repo,
		RepoOrder: us.RepoOrder,
		AuthRepo:  us.AuthRepo,
		Chat:      us.Chat,
		Business:  us.Business,
	}
}
