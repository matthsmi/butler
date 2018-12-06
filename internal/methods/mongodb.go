/*
Copyright 2017 Adobe. All rights reserved.
This file is licensed to you under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License. You may obtain a copy
of the License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
OF ANY KIND, either express or implied. See the License for the specific language
governing permissions and limitations under the License.
*/

package methods

import (
	"errors"
	"net/url"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	log "github.com/sirupsen/logrus"
)

type MongoDBOpts struct {
	URL *url.URL
	InsecureSkipVerifo bool
	Password string
	Username string
}

type MongoDBMethod struct {
}

func NewMongoDBMethod(manager *string, entry *string) (Method, error) {
	return MongoDBMethod{}, nil
}

func NewMongoDBMethodWithOptions(opts *MongoDBOpts) (Method, error) {
	// stegen finish
	return MongoDBMethod{}, nil
}

func (m MongoDBMethod) Get(u *url.URL) (*Response, error) {
	var (
		result *Response
	)
	log.Debugf("u=%#v", u)
	return result, errors.New("MongoDB method handler is not very useful, right now")
}
