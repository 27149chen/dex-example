/*
Copyright 2021 The KodeRover Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rest

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type engine struct {
	*gin.Engine

	mode string
}

func NewEngine() *engine {
	s := &engine{mode: "debug"}

	gin.SetMode(s.mode)

	s.injectMiddlewares()
	s.injectRouters()

	return s
}

func (s *engine) injectMiddlewares() {
	g := gin.New()
	defer func() {
		s.Engine = g
	}()

	if s.mode == gin.TestMode {
		return
	}

	g.Use(gin.Recovery())
}

func (s *engine) injectRouters() {
	g := s.Engine

	g.Any("/*path", func(c *gin.Context) {
		path := c.Param("path")
		if !strings.HasPrefix(path, "/api/users") {
			c.String(200, "OK")
			return
		}
		c.String(401, "Not authorized")
	})

	s.Engine = g
}
