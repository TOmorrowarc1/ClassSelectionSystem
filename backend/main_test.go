package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TOmorrowarc1/ClassSelectionSystem/account"
	"github.com/TOmorrowarc1/ClassSelectionSystem/course"
	"github.com/TOmorrowarc1/ClassSelectionSystem/privilege"
)

// setupTestServer initializes all subsystems for a clean test environment.
// This is crucial for making tests independent and repeatable.
func setupTestServer() {
	// These init functions create fresh, in-memory maps for each test run,
	// ensuring no data leaks between tests.
	account.InitAccountSystem()
	course.InitCourseSystem()
	privilege.InitPrivilegeSystem()
}

// Helper function to create a JSON request body for our API.
func createAPIRequestBody(action string, token string, params interface{}) *bytes.Buffer {
	// Using json.RawMessage for parameters to match the main router's logic
	paramBytes, _ := json.Marshal(params)
	requestBody := map[string]interface{}{
		"action":     action,
		"token":      token,
		"parameters": json.RawMessage(paramBytes),
	}
	bodyBytes, _ := json.Marshal(requestBody)
	return bytes.NewBuffer(bodyBytes)
}

// TestLoginAndAuthFlow covers the fundamental authentication process.
func TestLoginAndAuthFlow(t *testing.T) {
	setupTestServer()

	// The default admin user is created by InitAccountSystem
	loginParams := map[string]string{
		"name":     "admin",
		"password": "123456",
	}

	body := createAPIRequestBody("LogIn", "", loginParams)
	req := httptest.NewRequest(http.MethodPost, "/api", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute the request
	RequestRoute(rr, req)

	// Check the response
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", rr.Code)
	}

	var response struct {
		Token   string `json:"authToken"`
		Message string `json:"errorMessage"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	if response.Message != "" {
		t.Fatalf("Login failed with message: %s", response.Message)
	}
	if response.Token == "" {
		t.Fatal("Expected a non-empty auth token, but got an empty one")
	}

	// Now, test an action that requires this token
	adminToken := response.Token
	body = createAPIRequestBody("GetAllUsersInfo", adminToken, nil)
	req = httptest.NewRequest(http.MethodPost, "/api", body)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	RequestRoute(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GetAllUsersInfo with valid token failed with status: %d", rr.Code)
	}

	var getUsersResponse struct {
		Users   []interface{} `json:"users"`
		Message string        `json:"errorMessage"`
	}
	json.NewDecoder(rr.Body).Decode(&getUsersResponse)
	if getUsersResponse.Message != "" {
		t.Errorf("Expected no error message, but got: %s", getUsersResponse.Message)
	}
	if len(getUsersResponse.Users) == 0 {
		t.Error("Expected to get users, but got an empty list")
	}
}

// TestAdminActions tests endpoints that require admin privileges.
func TestAdminActions(t *testing.T) {
	setupTestServer()

	// First, get an admin token
	adminLoginParams := map[string]string{"name": "admin", "password": "123456"}
	body := createAPIRequestBody("LogIn", "", adminLoginParams)
	req := httptest.NewRequest(http.MethodPost, "/api", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	RequestRoute(rr, req)
	var loginResponse struct{ Token string `json:"authToken"` }
	json.NewDecoder(rr.Body).Decode(&loginResponse)
	adminToken := loginResponse.Token

	t.Run("AdminCanRegisterUser", func(t *testing.T) {
		registerParams := map[string]interface{}{
			"userInfo": map[string]interface{}{
				"username": "newstudent",
				"password": "password123",
				"Identity_info": map[string]interface{}{
					"privilege": "student",
					"Class":     map[string]int{"grade": 1, "class": 1},
				},
			},
		}
		body := createAPIRequestBody("Register", adminToken, registerParams)
		req := httptest.NewRequest(http.MethodPost, "/api", body)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		RequestRoute(rr, req)

		var resp struct{ Message string `json:"errorMessage"` }
		json.NewDecoder(rr.Body).Decode(&resp)

		if resp.Message != "" {
			t.Errorf("Admin failed to register user: %s", resp.Message)
		}
	})

	t.Run("StudentCannotRegisterUser", func(t *testing.T) {
		// First, register a student and get their token
		account.Register(account.UserInfo{Uid: "student1", Password: "pw", Privilege: account.PrivilegeStudent})
		studentLoginParams := map[string]string{"name": "student1", "password": "pw"}
		body := createAPIRequestBody("LogIn", "", studentLoginParams)
		req := httptest.NewRequest(http.MethodPost, "/api", body)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		RequestRoute(rr, req)
		var studentLoginResp struct{ Token string `json:"authToken"` }
		json.NewDecoder(rr.Body).Decode(&studentLoginResp)
		studentToken := studentLoginResp.Token

		// Now, try to register another user with the student token
		registerParams := map[string]interface{}{ /* ... same params as above ... */ }
		body = createAPIRequestBody("Register", studentToken, registerParams)
		req = httptest.NewRequest(http.MethodPost, "/api", body)
		req.Header.Set("Content-Type", "application/json")
		rr = httptest.NewRecorder()
		RequestRoute(rr, req)

		var resp struct{ Message string `json:"errorMessage"` }
		json.NewDecoder(rr.Body).Decode(&resp)

		if resp.Message != "Permission denied" {
			t.Errorf("Expected 'Permission denied', but got: '%s'", resp.Message)
		}
	})
}

// TestStudentCourseSelectionFlow tests the student-specific actions.
func TestStudentCourseSelectionFlow(t *testing.T) {
	setupTestServer()

	// Setup: Admin creates and launches a course
	course.AddCourse("Test Course", "Test Teacher", 2)
	course.LaunchCourse("Test Course")

	// Setup: Create a student user
	account.Register(account.UserInfo{Uid: "student_select", Password: "pw", Privilege: account.PrivilegeStudent})

	// Step 1: Student logs in
	studentLoginParams := map[string]string{"name": "student_select", "password": "pw"}
	body := createAPIRequestBody("LogIn", "", studentLoginParams)
	req := httptest.NewRequest(http.MethodPost, "/api", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	RequestRoute(rr, req)
	var loginResp struct{ Token string `json:"authToken"` }
	json.NewDecoder(rr.Body).Decode(&loginResp)
	studentToken := loginResp.Token

	// Step 2: Student selects the course
	selectParams := map[string]string{"courseName": "Test Course"}
	body = createAPIRequestBody("SelectCourse", studentToken, selectParams)
	req = httptest.NewRequest(http.MethodPost, "/api", body)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	RequestRoute(rr, req)

	var selectResp struct{ Message string `json:"errorMessage"` }
	json.NewDecoder(rr.Body).Decode(&selectResp)
	if selectResp.Message != "" {
		t.Fatalf("Student failed to select course: %s", selectResp.Message)
	}

	// Step 3: Student drops the course
	body = createAPIRequestBody("DropCourse", studentToken, nil) // No params needed for drop
	req = httptest.NewRequest(http.MethodPost, "/api", body)
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	RequestRoute(rr, req)

	var dropResp struct{ Message string `json:"errorMessage"` }
	json.NewDecoder(rr.Body).Decode(&dropResp)
	if dropResp.Message != "" {
		t.Fatalf("Student failed to drop course: %s", dropResp.Message)
	}
}