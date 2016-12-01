// Copyright (c) 2016 Readium Foundation
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation and/or
//    other materials provided with the distribution.
// 3. Neither the name of the organization nor the names of its contributors may be
//    used to endorse or promote products derived from this software without specific
//    prior written permission
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package webuser

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/readium/readium-lcp-server/api"
)

//ErrNotFound error trown when user is not found
var ErrNotFound = errors.New("User not found")

// WebUser interface for user db interaction
type WebUser interface {
	Get(id int64) (User, error)
	GetByEmail(email string) (User, error)
	Add(c User) error
	Update(c User) error
}

//User struct defines a user
type User struct {
	UserID   int64  `json:"userID"`
	Alias    string `json:"alias"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

//DecodeJSONUser transforms a json string to a User struct
func DecodeJSONUser(r *http.Request, user *User) error {
	var dec *json.Decoder

	if ctype := r.Header["Content-Type"]; len(ctype) > 0 && ctype[0] == api.ContentType_JSON {
		buf := bytes.NewBufferString(r.PostFormValue("data"))
		dec = json.NewDecoder(buf)
	} else {
		dec = json.NewDecoder(r.Body)
	}

	err := dec.Decode(&user)

	return err
}

type dbUser struct {
	db         *sql.DB
	getUser    *sql.Stmt
	getByEmail *sql.Stmt
	addUser    *sql.Stmt
	listUser   *sql.Stmt
}

func (user dbUser) Get(id int64) (User, error) {
	records, err := user.getUser.Query(id)
	defer records.Close()
	if records.Next() {
		var c User
		err = records.Scan(&c.UserID, &c.Alias, &c.Email, &c.Password)
		return c, err
	}

	return User{}, ErrNotFound
}

func (user dbUser) GetByEmail(email string) (User, error) {
	records, err := user.getByEmail.Query(email)
	defer records.Close()
	if records.Next() {
		var c User
		err = records.Scan(&c.UserID, &c.Alias, &c.Email, &c.Password)
		return c, err
	}

	return User{}, ErrNotFound
}

func (user dbUser) Add(newUser User) error {
	add, err := user.db.Prepare("INSERT INTO user (user_id,alias,email,password) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer add.Close()
	_, err = add.Exec(newUser.UserID, newUser.Alias, newUser.Email, newUser.Password)
	return err
}

func (user dbUser) Update(changedUser User) error {
	add, err := user.db.Prepare("UPDATE user SET alias=? , email=?, password=? WHERE user_id=?")
	if err != nil {
		return err
	}
	defer add.Close()
	_, err = add.Exec(changedUser.Alias, changedUser.Email, changedUser.Password, changedUser.UserID)
	return err
}

//Open  returns a WebUser interface (db interaction)
func Open(db *sql.DB) (i WebUser, err error) {
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS user (
	user_id integer NOT NULL, 
	alias varchar(64) NOT NULL, 
	email varchar(64) NOT NULL, 
	password varchar(64) NOT NULL, 
	constraint pk_user  primary key(user_id)
	)`)
	if err != nil {
		return
	}
	get, err := db.Prepare("SELECT user_id,alias,email,password FROM user WHERE user_id = ? LIMIT 1")
	if err != nil {
		return
	}
	getByEmail, err := db.Prepare("SELECT user_id,alias,email,password FROM user WHERE email = ? LIMIT 1")
	if err != nil {
		return
	}
	i = dbUser{db, get, getByEmail, nil, nil}
	return
}
