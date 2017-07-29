package auth

import (
	"testing"
	"time"
	"ventose.cc/data"
	"ventose.cc/portal/serverold/portal"
	"ventose.cc/tools"
)

func getAuthBackend(t *testing.T) *AuthBackend {
	ab := NewAuthBackend()
	p := portal.NewPortal()
	storageConfig := new(data.InitialConfiguration)
	storageConfig.AuthKey = tools.GetRandomAsciiString(10)
	storageConfig.ConfigPath = "/tmp/ledis.cfg"
	storageConfig.DbPath = "/tmp/data"
	storageConfig.MaxDatabases = 12
	storageConfig.PathAccessMode = 0770
	p.LoadStorage(storageConfig)
	err := p.ServeStorage()
	if err != nil {
		t.Fatal(err)
	}
	ab.ConnectStorage(p)
	ab.Run()
	return ab
}

func TestNewAuthBackend(t *testing.T) {
	//ab.R
	ab := getAuthBackend(t)

	u := new(User)
	u.Login = "test@test.de"
	u.Name = "Max Musterman"
	u.Created = time.Now()
	u.PassHash = tools.GetRandomBytes(52)
	//u.Id		= make([]byte, data.INDEX_SIZE)
	ab.AddChannel <- *u
	resp := <-ab.Response
	if resp.Error != nil {
		t.Fatal(resp)
	}

	if len(resp.UserId) < 1 {
		t.Errorf("Add did not return Userid")
	}
	u.Id = resp.UserId

	aur := new(AuthRequest)
	aur.Id = u.Id
	aur.Hash = u.PassHash

	ab.AuthChannel <- *aur

	ar := <-ab.Response
	if ar.Error != nil {
		t.Fatal(ar.Error)
	}

	if !ar.Authorized {
		t.Error("New User could not been Authorized by ID and Hash")
	}

	u.Login = "test2@test.de"
	newHash := tools.GetRandomBytes(24)
	u.PassHash = newHash

	ab.UpdateChannel <- *u
	uresp := <-ab.Response
	if uresp.Error != nil {
		t.Fatal(uresp)
	}
	ab.AuthChannel <- *aur

	uar := <-ab.Response
	if uar.Error != nil {
		t.Fatal(uar.Error)
	}

	if uar.Authorized {
		t.Error("Authorization succeded after update with old Hash")
	}

	aur.Hash = newHash
	aur.Id = nil
	aur.Login = u.Login

	ab.AuthChannel <- *aur

	u2ar := <-ab.Response
	if u2ar.Error != nil {
		t.Fatal(u2ar.Error)
	}

	if !u2ar.Authorized {
		t.Error("Updated User could not been Authorized by Login and Hash")
	}

	u.Id = u2ar.UserId
	ab.DelChannel <- *u

	dr := <-ab.Response

	if dr.Error != nil {
		t.Fatal(dr.Error)
	}

	ab.AuthChannel <- *aur
	ar = <-ab.Response
	if ar.Error != nil {
		t.Fatal(ar.Error)
	}

	if ar.Authorized {
		t.Fatal("Authorization succeded with deleted User")
	}
}
