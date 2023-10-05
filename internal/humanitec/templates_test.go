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

func TestMapVar(t *testing.T) {
	var meta = score.WorkloadMeta{
		Name: "test-name",
	}

	var resources = score.ResourcesSpecs{
		"env": score.ResourceSpec{
			Type: "environment",
		},
		"db": score.ResourceSpec{
			Type: "postgres",
		},
		"dns": score.ResourceSpec{
			Type: "dns",
		},
		"service-a": score.ResourceSpec{
			Type: "service",
		},
	}

	var ext = extensions.HumanitecResourcesSpecs{
		"dns": {Scope: "shared"},
	}

	ctx, err := buildContext(meta, resources, ext)
	assert.NoError(t, err)

	assert.Equal(t, "", ctx.mapVar(""))
	assert.Equal(t, "$", ctx.mapVar("$"))

	assert.Equal(t, "test-name", ctx.mapVar("metadata.name"))
	assert.Equal(t, "${metadata.name.nil}", ctx.mapVar("metadata.name.nil"))
	assert.Equal(t, "${metadata.nil}", ctx.mapVar("metadata.nil"))

	assert.Equal(t, "${values.DEBUG}", ctx.mapVar("resources.env.DEBUG"))

	assert.Equal(t, "externals.db", ctx.mapVar("resources.db"))
	assert.Equal(t, "${externals.db.host}", ctx.mapVar("resources.db.host"))
	assert.Equal(t, "${externals.db.port}", ctx.mapVar("resources.db.port"))
	assert.Equal(t, "${externals.db.name}", ctx.mapVar("resources.db.name"))
	assert.Equal(t, "${externals.db.name.nil}", ctx.mapVar("resources.db.name.nil"))
	assert.Equal(t, "${externals.db.nil}", ctx.mapVar("resources.db.nil"))
	assert.Equal(t, "${modules.service-a.service.name}", ctx.mapVar("resources.service-a.name"))
	assert.Equal(t, "${modules.service-a.service.port}", ctx.mapVar("resources.service-a.port"))
	assert.Equal(t, "${resources.nil}", ctx.mapVar("resources.nil"))
	assert.Equal(t, "${nil.db.name}", ctx.mapVar("nil.db.name"))
}

func TestEscape(t *testing.T) {
	var meta = score.WorkloadMeta{
		Name: "test-name",
	}

	var resources = score.ResourcesSpecs{
		"env": score.ResourceSpec{
			Type: "environment",
		},
		"db": score.ResourceSpec{
			Type: "postgres",
		},
		"dns": score.ResourceSpec{
			Type: "dns",
		},
		"service-a": score.ResourceSpec{
			Type: "service",
		},
	}

	var ext = extensions.HumanitecResourcesSpecs{
		"dns": {Scope: "shared"},
	}

	ctx, err := buildContext(meta, resources, ext)
	assert.NoError(t, err)

	assert.Equal(t, "", ctx.Escape(""))
	assert.Equal(t, "abc", ctx.Escape("abc"))
	assert.Equal(t, "abc $$ abc", ctx.Escape("abc $$ abc"))
	assert.Equal(t, "$abc", ctx.Escape("$abc"))
	assert.Equal(t, "$${abc}", ctx.Escape("$${abc}"))

	assert.Equal(t, "The name is '$\\{metadata.name}'", ctx.Escape("The name is '${metadata.name}'"))
	assert.Equal(t, "The name is '$\\{metadata.nil}'", ctx.Escape("The name is '${metadata.nil}'"))

	assert.Equal(t, "resources.env.DEBUG", ctx.Escape("resources.env.DEBUG"))

	assert.Equal(t, "$\\{resources.db}", ctx.Escape("${resources.db}"))
	assert.Equal(t,
		"postgresql://$\\{resources.db.user}:$\\{resources.db.password}@$\\{resources.db.host}:$\\{resources.db.port}/$\\{resources.db.name}",
		ctx.Escape("postgresql://${resources.db.user}:${resources.db.password}@${resources.db.host}:${resources.db.port}/${resources.db.name}"))
}

func TestSubstitute(t *testing.T) {
	var meta = score.WorkloadMeta{
		Name: "test-name",
	}

	var resources = score.ResourcesSpecs{
		"env": score.ResourceSpec{
			Type: "environment",
		},
		"db": score.ResourceSpec{
			Type: "postgres",
		},
		"dns": score.ResourceSpec{
			Type: "dns",
		},
		"service-a": score.ResourceSpec{
			Type: "service",
		},
	}

	var ext = extensions.HumanitecResourcesSpecs{
		"dns": {Scope: "shared"},
	}

	ctx, err := buildContext(meta, resources, ext)
	assert.NoError(t, err)

	assert.Equal(t, "", ctx.Substitute(""))
	assert.Equal(t, "abc", ctx.Substitute("abc"))
	assert.Equal(t, "abc $ abc", ctx.Substitute("abc $$ abc"))
	assert.Equal(t, "$abc", ctx.Substitute("$abc"))
	assert.Equal(t, "${abc}", ctx.Substitute("$${abc}"))

	assert.Equal(t, "The name is 'test-name'", ctx.Substitute("The name is '${metadata.name}'"))
	assert.Equal(t, "The name is '${metadata.nil}'", ctx.Substitute("The name is '${metadata.nil}'"))

	assert.Equal(t, "resources.env.DEBUG", ctx.Substitute("resources.env.DEBUG"))

	assert.Equal(t, "externals.db", ctx.Substitute("${resources.db}"))
	assert.Equal(t,
		"postgresql://${externals.db.user}:${externals.db.password}@${externals.db.host}:${externals.db.port}/${externals.db.name}",
		ctx.Substitute("postgresql://${resources.db.user}:${resources.db.password}@${resources.db.host}:${resources.db.port}/${resources.db.name}"))
}

func TestSubstituteAll(t *testing.T) {
	var meta = score.WorkloadMeta{
		Name: "test-name",
	}

	var resources = score.ResourcesSpecs{
		"env": score.ResourceSpec{
			Type: "environment",
		},
		"db": score.ResourceSpec{
			Type: "postgres",
		},
		"dns": score.ResourceSpec{
			Type: "dns",
		},
		"service-a": score.ResourceSpec{
			Type: "service",
		},
	}

	var ext = extensions.HumanitecResourcesSpecs{
		"dns": {Scope: "shared"},
	}

	ctx, err := buildContext(meta, resources, ext)
	assert.NoError(t, err)

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

	assert.Equal(t, expected, ctx.SubstituteAll(source))
}
