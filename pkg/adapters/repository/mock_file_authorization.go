package repository

type MockFileAuthorization struct{}

func NewMockFileAuthorization() *MockFileAuthorization {
	return &MockFileAuthorization{}
}

func (m *MockFileAuthorization) CanUploadFile(userID, fileType, linkedResourceType, linkedResourceID string) (bool, error) {
	return true, nil
}

func (m *MockFileAuthorization) CanReadFile(userID, fileID string) (bool, error) {
	return true, nil
}

func (m *MockFileAuthorization) CanDeleteFile(userID, fileID string) (bool, error) {
	return true, nil
}

func (m *MockFileAuthorization) CreateFileAuthorization(fileID, fileType, linkedResourceID, linkedResourceType string) error {
	return nil
}

func (m *MockFileAuthorization) RemoveFileAuthorization(fileID, fileType, linkedResourceID, linkedResourceType string) error {
	return nil
}
