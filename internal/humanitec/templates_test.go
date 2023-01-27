/*
Apache Score
Copyright 2022 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package humanitec

import (
	"testing"

	score "github.com/score-spec/score-go/types"
	assert "github.com/stretchr/testify/assert"

	extensions "github.com/score-spec/score-humanitec/internal/humanitec/extensions"
)

func TestBuildContext(t *testing.T) {
	var meta = score.WorkloadMeta{
		Name: "test-name",
	}

	var resources = score.ResourcesSpecs{
		"env": score.ResourceSpec{
			Type: "environment",
			Properties: map[string]score.ResourcePropertySpec{
				"DEBUG": {Required: false, Default: true},
			},
		},
		"db": score.ResourceSpec{
			Type: "postgres",
			Properties: map[string]score.ResourcePropertySpec{
				"host": {Required: true, Default: "."},
				"port": {Required: true, Default: "5342"},
				"name": {Required: true},
			},
		},
		"dns": score.ResourceSpec{
			Type: "dns",
			Properties: map[string]score.ResourcePropertySpec{
				"domain": {},
			},
		},
		"service-a": score.ResourceSpec{
			Type: "service",
			Properties: map[string]score.ResourcePropertySpec{
				"name": {},
				"port": {},
			},
		},
	}

	var ext = extensions.HumanitecResourcesSpecs{
		"dns": {Scope: "shared"},
	}

	context, err := buildContext(meta, resources, ext)
	assert.NoError(t, err)

	assert.Equal(t, templatesContext{
		"metadata.name": "test-name",

		"resources.env":       "values",
		"resources.env.DEBUG": "${values.DEBUG}",

		"resources.db":      "externals.db",
		"resources.db.host": "${externals.db.host}",
		"resources.db.port": "${externals.db.port}",
		"resources.db.name": "${externals.db.name}",

		"resources.dns":        "shared.dns",
		"resources.dns.domain": "${shared.dns.domain}",

		"resources.service-a":      "modules.service-a",
		"resources.service-a.name": "${modules.service-a.service.name}",
		"resources.service-a.port": "${modules.service-a.service.port}",
	}, context)
}

func TestMapVar(t *testing.T) {
	var context = templatesContext{
		"metadata.name": "test-name",

		"resources.env":       "values",
		"resources.env.DEBUG": "${values.DEBUG}",

		"resources.db":      "externals.db",
		"resources.db.host": "${externals.db.host}",
		"resources.db.port": "${externals.db.port}",
		"resources.db.name": "${externals.db.name}",

		"resources.dns":        "shared.dns",
		"resources.dns.domain": "${shared.dns.domain}",

		"resources.service-a":      "modules.service-a",
		"resources.service-a.name": "${modules.service-a.service.name}",
		"resources.service-a.port": "${modules.service-a.service.port}",
	}

	assert.Equal(t, "", context.mapVar(""))
	assert.Equal(t, "$", context.mapVar("$"))

	assert.Equal(t, "test-name", context.mapVar("metadata.name"))
	assert.Equal(t, "${metadata.name.nil}", context.mapVar("metadata.name.nil"))
	assert.Equal(t, "${metadata.nil}", context.mapVar("metadata.nil"))

	assert.Equal(t, "${values.DEBUG}", context.mapVar("resources.env.DEBUG"))

	assert.Equal(t, "externals.db", context.mapVar("resources.db"))
	assert.Equal(t, "${externals.db.host}", context.mapVar("resources.db.host"))
	assert.Equal(t, "${externals.db.port}", context.mapVar("resources.db.port"))
	assert.Equal(t, "${externals.db.name}", context.mapVar("resources.db.name"))
	assert.Equal(t, "${resources.db.name.nil}", context.mapVar("resources.db.name.nil"))
	assert.Equal(t, "${resources.db.nil}", context.mapVar("resources.db.nil"))
	assert.Equal(t, "${modules.service-a.service.name}", context.mapVar("resources.service-a.name"))
	assert.Equal(t, "${modules.service-a.service.port}", context.mapVar("resources.service-a.port"))
	assert.Equal(t, "${resources.nil}", context.mapVar("resources.nil"))
	assert.Equal(t, "${nil.db.name}", context.mapVar("nil.db.name"))
}

func TestSubstitute(t *testing.T) {
	var context = templatesContext{
		"metadata.name": "test-name",

		"resources.env":       "values",
		"resources.env.DEBUG": "${values.DEBUG}",

		"resources.db":      "externals.db",
		"resources.db.host": "${externals.db.host}",
		"resources.db.port": "${externals.db.port}",
		"resources.db.name": "${externals.db.name}",

		"resources.dns":        "shared.dns",
		"resources.dns.domain": "${shared.dns.domain}",

		"resources.service-a":      "modules.service-a",
		"resources.service-a.name": "${modules.service-a.service.name}",
		"resources.service-a.port": "${modules.service-a.service.port}",
	}

	assert.Equal(t, "", context.Substitute(""))
	assert.Equal(t, "abc", context.Substitute("abc"))
	assert.Equal(t, "abc $ abc", context.Substitute("abc $$ abc"))
	assert.Equal(t, "${abc}", context.Substitute("$${abc}"))

	assert.Equal(t, "The name is 'test-name'", context.Substitute("The name is '${metadata.name}'"))
	assert.Equal(t, "The name is '${metadata.nil}'", context.Substitute("The name is '${metadata.nil}'"))

	assert.Equal(t, "resources.env.DEBUG", context.Substitute("resources.env.DEBUG"))

	assert.Equal(t, "externals.db", context.Substitute("${resources.db}"))
	assert.Equal(t,
		"postgresql://${resources.db.user}:${resources.db.password}@${externals.db.host}:${externals.db.port}/${externals.db.name}",
		context.Substitute("postgresql://${resources.db.user}:${resources.db.password}@${resources.db.host}:${resources.db.port}/${resources.db.name}"))
}

func TestSubstituteAll(t *testing.T) {
	var context = templatesContext{
		"metadata.name": "test-name",

		"resources.env":       "values",
		"resources.env.DEBUG": "${values.DEBUG}",

		"resources.db":      "externals.db",
		"resources.db.host": "${externals.db.host}",
		"resources.db.port": "${externals.db.port}",
		"resources.db.name": "${externals.db.name}",

		"resources.dns":        "shared.dns",
		"resources.dns.domain": "${shared.dns.domain}",

		"resources.service-a":      "modules.service-a",
		"resources.service-a.name": "${modules.service-a.service.name}",
		"resources.service-a.port": "${modules.service-a.service.port}",
	}

	var source = map[string]interface{}{
		"api": map[string]interface{}{
			"${resources.service-a.name}": map[string]interface{}{
				"url":   "http://${resources.dns.domain}",
				"port":  "${resources.service-a.port}",
				"retry": 10,
			},
		},
		"DEBUG": "${resources.env.DEBUG}",
	}

	var expected = map[string]interface{}{
		"api": map[string]interface{}{
			"${modules.service-a.service.name}": map[string]interface{}{
				"url":   "http://${shared.dns.domain}",
				"port":  "${modules.service-a.service.port}",
				"retry": 10,
			},
		},
		"DEBUG": "${values.DEBUG}",
	}

	assert.Equal(t, expected, context.SubstituteAll(source))
}
