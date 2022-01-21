package provider

import gogit "github.com/go-git/go-git/v5"

func getRepository(dir string) (*gogit.Repository, *gogit.Worktree, error) {
	// Open existing repository
	repo, err := gogit.PlainOpen(dir)
	if err != nil {
		return nil, nil, err
	}

	// Get the current worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return repo, nil, err
	}

	return repo, worktree, nil
}
