package repository

type MockFileAuthorization struct{}

func NewMockFileAuthorization() *MockFileAuthorization {
	return &MockFileAuthorization{}
}

func (m *MockFileAuthorization) AuthorizeUploadFile(userID, fileType, linkedResourceType, linkedResourceID string) (bool, error) {
	return true, nil
}

func (m *MockFileAuthorization) AuthorizeReadFile(userID, fileID string) (bool, error) {
	return true, nil
}

func (m *MockFileAuthorization) AuthorizeDeleteFile(userID, fileID string) (bool, error) {
	return true, nil
}
