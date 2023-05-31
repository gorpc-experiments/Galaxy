package service

import "github.com/gorpc-experiments/galaxy/src/domain"

type Galaxy struct {
	ServiceLibrary []domain.Module
}

func NewGalaxy() *Galaxy {
	galaxy := new(Galaxy)

	galaxy.ServiceLibrary = []domain.Module{}

	return galaxy
}