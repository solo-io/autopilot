package model

import (
	"fmt"
	"strings"

	"github.com/gertd/go-pluralize"
)

var (
	// function for determining the relative path of generated api types package
	TypesRelativePath = func(kind, version string) string {
		c := pluralize.NewClient()
		return fmt.Sprintf("pkg/apis/%v/%v", strings.ToLower(c.Plural(kind)), version)
	}

	// function for determining the relative path of generated schduler package
	SchedulerRelativePath = "pkg/scheduler"

	// function for determining the relative path of generated finalizer package
	FinalizerRelativePath = "pkg/finalizer"

	// function for determining the relative path of generated parameters package
	ParametersRelativePath = "pkg/parameters"

	// function for determining the relative path of generated metrics package
	MetricsRelativePath = "pkg/metrics"
)
