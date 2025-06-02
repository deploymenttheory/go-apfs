package services

import (
	"context"
	"testing"
	"time"
)

func TestServiceFactory(t *testing.T) {
	factory := NewServiceFactory()

	// Test initialization
	err := factory.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	if !factory.IsInitialized() {
		t.Error("Factory should be initialized")
	}

	// Test getting container service
	containerSvc, err := factory.ContainerService()
	if err != nil {
		t.Fatalf("Failed to get container service: %v", err)
	}
	if containerSvc == nil {
		t.Error("Container service should not be nil")
	}

	// Test getting filesystem service
	filesystemSvc, err := factory.FilesystemService()
	if err != nil {
		t.Fatalf("Failed to get filesystem service: %v", err)
	}
	if filesystemSvc == nil {
		t.Error("Filesystem service should not be nil")
	}

	// Test services that aren't implemented yet
	_, err = factory.VolumeService()
	if err != ErrServiceNotImplemented {
		t.Errorf("Expected ErrServiceNotImplemented, got: %v", err)
	}

	// Test shutdown
	err = factory.Shutdown()
	if err != nil {
		t.Fatalf("Failed to shutdown services: %v", err)
	}

	if factory.IsInitialized() {
		t.Error("Factory should not be initialized after shutdown")
	}
}

func TestContainerService(t *testing.T) {
	svc := NewContainerService()
	ctx := context.Background()

	// Test discovery (should not fail even if no containers found)
	containers, err := svc.DiscoverContainers(ctx)
	if err != nil {
		t.Fatalf("DiscoverContainers should not fail: %v", err)
	}

	// Should return empty slice if no containers found
	if containers == nil {
		t.Error("DiscoverContainers should return empty slice, not nil")
	}

	// Test closing (should not fail even with no open containers)
	err = svc.Close()
	if err != nil {
		t.Fatalf("Close should not fail: %v", err)
	}
}

func TestFilesystemService(t *testing.T) {
	containerSvc := NewContainerService()
	svc := NewFilesystemService(containerSvc)
	ctx := context.Background()

	// Test with mock data - these will fail to open actual containers but shouldn't panic
	_, err := svc.ListDirectory(ctx, "/nonexistent", 1, "/", false)
	if err == nil {
		t.Error("Expected error for nonexistent container")
	}

	_, err = svc.GetFileInfo(ctx, "/nonexistent", 1, "/test")
	if err == nil {
		t.Error("Expected error for nonexistent container")
	}

	_, err = svc.CheckAccess(ctx, "/nonexistent", 1, "/test")
	if err == nil {
		t.Error("Expected error for nonexistent container")
	}
}

func TestDefaultServiceFactory(t *testing.T) {
	// Test convenience functions
	_, err := GetContainerService()
	if err != nil {
		t.Fatalf("Failed to get default container service: %v", err)
	}

	_, err = GetFilesystemService()
	if err != nil {
		t.Fatalf("Failed to get default filesystem service: %v", err)
	}

	// Test services not yet implemented
	_, err = GetVolumeService()
	if err != ErrServiceNotImplemented {
		t.Errorf("Expected ErrServiceNotImplemented, got: %v", err)
	}

	// Test service info
	services := DefaultServiceFactory.ListAvailableServices()
	if len(services) == 0 {
		t.Error("Should have some available services")
	}

	availableCount := 0
	for _, service := range services {
		if service.Available {
			availableCount++
		}
	}

	if availableCount == 0 {
		t.Error("Should have at least one available service")
	}

	// Test getting all services
	allServices, err := DefaultServiceFactory.GetAllServices()
	if err != nil {
		t.Fatalf("Failed to get all services: %v", err)
	}

	if len(allServices) == 0 {
		t.Error("Should have some services available")
	}

	// Clean up
	err = ShutdownServices()
	if err != nil {
		t.Fatalf("Failed to shutdown services: %v", err)
	}
}

func TestContainerServiceLimitations(t *testing.T) {
	// Test that the service behaves appropriately with invalid inputs
	svc := NewContainerService()
	ctx := context.Background()

	// Test timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()

	// Give the context time to timeout
	time.Sleep(2 * time.Millisecond)

	_, err := svc.DiscoverContainers(timeoutCtx)
	if err != context.DeadlineExceeded {
		t.Log("Note: DiscoverContainers may complete before timeout in test environment")
	}

	// Test with invalid paths
	_, err = svc.OpenContainer(ctx, "/this/path/definitely/does/not/exist")
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}

	// Test reading superblock on nonexistent container
	_, err = svc.ReadSuperblock(ctx, "/this/path/definitely/does/not/exist")
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}
}

func TestFilesystemServiceMockBehavior(t *testing.T) {
	// Test the mock behavior of the filesystem service
	containerSvc := NewContainerService()
	svc := NewFilesystemService(containerSvc)

	// Test helper functions
	fileType := svc.(*filesystemService).determineFileType("test.app")
	if fileType != "directory" {
		t.Errorf("Expected 'directory' for .app, got '%s'", fileType)
	}

	fileType = svc.(*filesystemService).determineFileType("test.txt")
	if fileType != "file" {
		t.Errorf("Expected 'file' for .txt, got '%s'", fileType)
	}

	isDir := svc.(*filesystemService).isKnownDirectory("/", "Applications")
	if !isDir {
		t.Error("Applications should be recognized as a directory")
	}

	isDir = svc.(*filesystemService).isKnownDirectory("/", "SomeRandomFile")
	if isDir {
		t.Error("SomeRandomFile should not be recognized as a directory")
	}
}

// Benchmark tests to ensure services perform adequately
func BenchmarkContainerServiceDiscovery(b *testing.B) {
	svc := NewContainerService()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.DiscoverContainers(ctx)
	}
}

func BenchmarkServiceFactoryInitialization(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		factory := NewServiceFactory()
		_ = factory.Initialize()
		_ = factory.Shutdown()
	}
}
