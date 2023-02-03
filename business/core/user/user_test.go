package user_test

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/core/user/stores/userdb"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/foundation/docker"
	"github.com/google/go-cmp/cmp"
)

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	m.Run()
}

func Test_User(t *testing.T) {
	log, db, teardown := dbtest.NewUnit(t, c, "testuser")
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		teardown()
	}()

	core := user.NewCore(userdb.NewStore(log, db))

	t.Log("Given the need to work with User records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single User.", testID)
		{
			ctx := context.Background()

			email, err := mail.ParseAddress("bill@ardanlabs.com")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse email: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse email.", dbtest.Success, testID)

			nu := user.NewUser{
				Name:            "Bill Kennedy",
				Email:           *email,
				Roles:           []string{user.RoleAdmin},
				Password:        "gophers",
				PasswordConfirm: "gophers",
			}

			usr, err := core.Create(ctx, nu)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create user.", dbtest.Success, testID)

			saved, err := core.QueryByID(ctx, usr.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user by ID: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve user by ID.", dbtest.Success, testID)

			if usr.DateCreated.UnixMilli() != saved.DateCreated.UnixMilli() {
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.DateCreated)
				t.Logf("\t\tTest %d:\tExp: %v", testID, usr.DateCreated)
				t.Logf("\t\tTest %d:\tDiff: %v", testID, saved.DateCreated.Sub(usr.DateCreated))
				t.Fatalf("\t%s\tTest %d:\tShould get back the same date created.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same date created.", dbtest.Success, testID)

			if usr.DateUpdated.UnixMilli() != saved.DateUpdated.UnixMilli() {
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.DateUpdated)
				t.Logf("\t\tTest %d:\tExp: %v", testID, usr.DateUpdated)
				t.Logf("\t\tTest %d:\tDiff: %v", testID, saved.DateUpdated.Sub(usr.DateUpdated))
				t.Fatalf("\t%s\tTest %d:\tShould get back the same date updated.", dbtest.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same date updated.", dbtest.Success, testID)

			usr.DateCreated = time.Time{}
			usr.DateUpdated = time.Time{}
			saved.DateCreated = time.Time{}
			saved.DateUpdated = time.Time{}

			if diff := cmp.Diff(usr, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same user. Diff:\n%s", dbtest.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same user.", dbtest.Success, testID)

			email, err = mail.ParseAddress("jacob@ardanlabs.com")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse email: %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse email.", dbtest.Success, testID)

			upd := user.UpdateUser{
				Name:  dbtest.StringPointer("Jacob Walker"),
				Email: email,
			}

			if _, err := core.Update(ctx, saved, upd); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update user.", dbtest.Success, testID)

			saved, err = core.QueryByEmail(ctx, *upd.Email)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve user by Email : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve user by Email.", dbtest.Success, testID)

			diff := usr.DateUpdated.Sub(saved.DateUpdated)
			if diff > 0 {
				t.Fatalf("Should have a larger DateUpdated : sav %v, usr %v, dif %v", saved.DateUpdated, usr.DateUpdated, diff)
			}

			if saved.Name != *upd.Name {
				t.Errorf("\t%s\tTest %d:\tShould be able to see updates to Name.", dbtest.Failed, testID)
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.Name)
				t.Logf("\t\tTest %d:\tExp: %v", testID, *upd.Name)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to Name.", dbtest.Success, testID)
			}

			if saved.Email != *upd.Email {
				t.Errorf("\t%s\tTest %d:\tShould be able to see updates to Email.", dbtest.Failed, testID)
				t.Logf("\t\tTest %d:\tGot: %v", testID, saved.Email)
				t.Logf("\t\tTest %d:\tExp: %v", testID, *upd.Email)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updates to Email.", dbtest.Success, testID)
			}

			if err := core.Delete(ctx, saved); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete user.", dbtest.Success, testID)

			_, err = core.QueryByID(ctx, saved.ID)
			if !errors.Is(err, user.ErrNotFound) {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve user : %s.", dbtest.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve user.", dbtest.Success, testID)
		}
	}
}
