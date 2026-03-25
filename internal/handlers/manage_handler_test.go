package handlers

import (
	"testing"

	"github.com/jchanning/gocase/internal/auth"
	"github.com/jchanning/gocase/internal/models"
)

func intPtr(v int) *int { return &v }

func TestCanManageTest_NilSessionReturnsFalse(t *testing.T) {
	h := &ManageHandler{}
	test := &models.Test{ID: 1, CreatedBy: intPtr(5)}
	if h.canManageTest(nil, test) {
		t.Fatal("expected canManageTest to return false for nil session")
	}
}

func TestCanManageTest_NilTestReturnsFalse(t *testing.T) {
	h := &ManageHandler{}
	session := &auth.SessionData{UserID: 1, Role: "teacher"}
	if h.canManageTest(session, nil) {
		t.Fatal("expected canManageTest to return false for nil test")
	}
}

func TestCanManageTest_AdminCanManageAnyTest(t *testing.T) {
	h := &ManageHandler{}
	session := &auth.SessionData{UserID: 99, Role: "admin"}
	test := &models.Test{ID: 1, CreatedBy: intPtr(5)} // different creator

	if !h.canManageTest(session, test) {
		t.Fatal("expected admin to be able to manage any test")
	}
}

func TestCanManageTest_TeacherCanManageOwnTest(t *testing.T) {
	h := &ManageHandler{}
	session := &auth.SessionData{UserID: 10, Role: "teacher"}
	test := &models.Test{ID: 1, CreatedBy: intPtr(10)}

	if !h.canManageTest(session, test) {
		t.Fatal("expected teacher to manage their own test")
	}
}

func TestCanManageTest_TeacherCannotManageOtherTest(t *testing.T) {
	h := &ManageHandler{}
	session := &auth.SessionData{UserID: 10, Role: "teacher"}
	test := &models.Test{ID: 1, CreatedBy: intPtr(20)} // owned by someone else

	if h.canManageTest(session, test) {
		t.Fatal("expected teacher to be denied access to another user's test")
	}
}

func TestCanManageTest_TeacherWithNilCreatedBy(t *testing.T) {
	h := &ManageHandler{}
	session := &auth.SessionData{UserID: 10, Role: "teacher"}
	test := &models.Test{ID: 1, CreatedBy: nil}

	// A test with no creator should not be manageable by a teacher.
	if h.canManageTest(session, test) {
		t.Fatal("expected teacher to be denied access to a test with nil CreatedBy")
	}
}

func TestCanManageTest_StudentCannotManageTest(t *testing.T) {
	h := &ManageHandler{}
	session := &auth.SessionData{UserID: 10, Role: "student"}
	test := &models.Test{ID: 1, CreatedBy: intPtr(10)} // even their own user ID

	if h.canManageTest(session, test) {
		t.Fatal("expected student to be denied manage access")
	}
}
