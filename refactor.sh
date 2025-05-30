#!/bin/bash

make_folder() {
    mkdir -p internal/{models,handler,repo,service}
    mkdir -p pkg/{config,logger}
    mkdir -p cmd/server
    
    touch internal/models/models.go
    touch internal/handler/handler.go
    touch internal/repo/repo.go
    touch internal/service/service.go
    touch pkg/config/config.go
    touch pkg/logger/logger.go
    touch cmd/server/main.go
    
    echo "Folder structure created successfully!"
    echo "Created directories:"
    echo "- internal/models (data models and structs)"
    echo "- internal/handler (HTTP handlers)"
    echo "- internal/repo (data access layer)"
    echo "- internal/service (business logic)"
    echo "- pkg/config (configuration management)"
    echo "- pkg/logger (logging utilities)"
    echo "- cmd/server (application entry point)"
}

make_folder