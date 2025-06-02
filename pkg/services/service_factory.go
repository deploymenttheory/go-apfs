package services

import (
	"fmt"
	"sync"
)

// ServiceFactory provides a centralized way to create and manage APFS services
type ServiceFactory struct {
	containerService  ContainerService
	filesystemService FilesystemService
	analysisService   AnalysisService
	extractionService ExtractionService
	volumeService     VolumeService
	mu                sync.RWMutex
	initialized       bool
}

// NewServiceFactory creates a new service factory instance
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

// Initialize initializes all services with their dependencies
func (sf *ServiceFactory) Initialize() error {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	if sf.initialized {
		return nil
	}

	// Create container service first (it's the foundation)
	sf.containerService = NewContainerService()

	// Create filesystem service (depends on container service)
	sf.filesystemService = NewFilesystemService(sf.containerService)

	// Create other services would go here when implemented
	// sf.volumeService = NewVolumeService(sf.containerService)
	// sf.analysisService = NewAnalysisService(sf.containerService, sf.volumeService)
	// sf.extractionService = NewExtractionService(sf.containerService, sf.filesystemService)

	sf.initialized = true
	return nil
}

// ContainerService returns the container service instance
func (sf *ServiceFactory) ContainerService() (ContainerService, error) {
	sf.mu.RLock()
	defer sf.mu.RUnlock()

	if !sf.initialized {
		sf.mu.RUnlock()
		if err := sf.Initialize(); err != nil {
			return nil, err
		}
		sf.mu.RLock()
	}

	return sf.containerService, nil
}

// FilesystemService returns the filesystem service instance
func (sf *ServiceFactory) FilesystemService() (FilesystemService, error) {
	sf.mu.RLock()
	defer sf.mu.RUnlock()

	if !sf.initialized {
		sf.mu.RUnlock()
		if err := sf.Initialize(); err != nil {
			return nil, err
		}
		sf.mu.RLock()
	}

	return sf.filesystemService, nil
}

// VolumeService returns the volume service instance
func (sf *ServiceFactory) VolumeService() (VolumeService, error) {
	sf.mu.RLock()
	defer sf.mu.RUnlock()

	if !sf.initialized {
		sf.mu.RUnlock()
		if err := sf.Initialize(); err != nil {
			return nil, err
		}
		sf.mu.RLock()
	}

	// TODO: Implement VolumeService
	return nil, ErrServiceNotImplemented
}

// AnalysisService returns the analysis service instance
func (sf *ServiceFactory) AnalysisService() (AnalysisService, error) {
	sf.mu.RLock()
	defer sf.mu.RUnlock()

	if !sf.initialized {
		sf.mu.RUnlock()
		if err := sf.Initialize(); err != nil {
			return nil, err
		}
		sf.mu.RLock()
	}

	// TODO: Implement AnalysisService
	return nil, ErrServiceNotImplemented
}

// ExtractionService returns the extraction service instance
func (sf *ServiceFactory) ExtractionService() (ExtractionService, error) {
	sf.mu.RLock()
	defer sf.mu.RUnlock()

	if !sf.initialized {
		sf.mu.RUnlock()
		if err := sf.Initialize(); err != nil {
			return nil, err
		}
		sf.mu.RLock()
	}

	// TODO: Implement ExtractionService
	return nil, ErrServiceNotImplemented
}

// Shutdown gracefully shuts down all services
func (sf *ServiceFactory) Shutdown() error {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	if !sf.initialized {
		return nil
	}

	// Close container service (this will close all open containers)
	if sf.containerService != nil {
		if err := sf.containerService.Close(); err != nil {
			return err
		}
	}

	// Reset all services
	sf.containerService = nil
	sf.filesystemService = nil
	sf.analysisService = nil
	sf.extractionService = nil
	sf.volumeService = nil
	sf.initialized = false

	return nil
}

// IsInitialized returns whether the factory has been initialized
func (sf *ServiceFactory) IsInitialized() bool {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.initialized
}

// GetAllServices returns all available services
func (sf *ServiceFactory) GetAllServices() (map[string]interface{}, error) {
	if err := sf.Initialize(); err != nil {
		return nil, err
	}

	services := make(map[string]interface{})

	if containerSvc, err := sf.ContainerService(); err == nil {
		services["container"] = containerSvc
	}

	if filesystemSvc, err := sf.FilesystemService(); err == nil {
		services["filesystem"] = filesystemSvc
	}

	// Add other services as they're implemented
	// if volumeSvc, err := sf.VolumeService(); err == nil {
	//     services["volume"] = volumeSvc
	// }

	return services, nil
}

// ServiceInfo represents information about a service
type ServiceInfo struct {
	Name        string
	Description string
	Available   bool
	Version     string
}

// ListAvailableServices returns information about all available services
func (sf *ServiceFactory) ListAvailableServices() []ServiceInfo {
	services := []ServiceInfo{
		{
			Name:        "container",
			Description: "Container discovery, superblock reading, and basic container metadata analysis",
			Available:   true,
			Version:     "1.0.0",
		},
		{
			Name:        "filesystem",
			Description: "File and directory listing, inode reading, and basic filesystem navigation",
			Available:   true,
			Version:     "1.0.0",
		},
		{
			Name:        "volume",
			Description: "Volume enumeration, metadata reading, and basic volume information extraction",
			Available:   false,
			Version:     "1.0.0",
		},
		{
			Name:        "analysis",
			Description: "Deep APFS structure analysis including B-tree analysis and filesystem health assessment",
			Available:   false,
			Version:     "1.0.0",
		},
		{
			Name:        "extraction",
			Description: "File and directory extraction with metadata preservation and integrity verification",
			Available:   false,
			Version:     "1.0.0",
		},
	}

	return services
}

// Common errors
var (
	ErrServiceNotImplemented = fmt.Errorf("service not yet implemented")
	ErrServiceNotAvailable   = fmt.Errorf("service not available")
)

// DefaultServiceFactory is the default global service factory instance
var DefaultServiceFactory = NewServiceFactory()

// Convenience functions for accessing services through the default factory

// GetContainerService returns the default container service
func GetContainerService() (ContainerService, error) {
	return DefaultServiceFactory.ContainerService()
}

// GetFilesystemService returns the default filesystem service
func GetFilesystemService() (FilesystemService, error) {
	return DefaultServiceFactory.FilesystemService()
}

// GetVolumeService returns the default volume service
func GetVolumeService() (VolumeService, error) {
	return DefaultServiceFactory.VolumeService()
}

// GetAnalysisService returns the default analysis service
func GetAnalysisService() (AnalysisService, error) {
	return DefaultServiceFactory.AnalysisService()
}

// GetExtractionService returns the default extraction service
func GetExtractionService() (ExtractionService, error) {
	return DefaultServiceFactory.ExtractionService()
}

// InitializeServices initializes all services using the default factory
func InitializeServices() error {
	return DefaultServiceFactory.Initialize()
}

// ShutdownServices shuts down all services using the default factory
func ShutdownServices() error {
	return DefaultServiceFactory.Shutdown()
}
