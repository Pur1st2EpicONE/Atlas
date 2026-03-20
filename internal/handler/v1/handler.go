package v1

import (
	"Atlas/internal/config"
	"Atlas/internal/service"
)

type Handler struct {
	config  config.Server
	service service.Service
}

func NewHandler(config config.Server, service service.Service) *Handler {
	return &Handler{config: config, service: service}
}
