package gateway

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/TykTechnologies/tyk/apidef"
	"github.com/TykTechnologies/tyk/test"
)

var testJsonSchema = `{
    "title": "Person",
    "type": "object",
    "properties": {
        "firstName": {
            "type": "string"
        },
        "lastName": {
            "type": "string"
        },
        "age": {
            "description": "Age in years",
            "type": "integer",
            "minimum": 0
        },
		"objs":{
			"enum":["a","b","c"],
			"type":"string"
		}
    },
    "required": ["firstName", "lastName"]
}`

func testPrepareValidateJSONSchema() {
	BuildAndLoadAPI(func(spec *APISpec) {
		UpdateAPIVersion(spec, "v1", func(v *apidef.VersionInfo) {
			json.Unmarshal([]byte(`[
				{
					"path": "/v",
					"method": "POST",
					"schema": `+testJsonSchema+`
				}
			]`), &v.ExtendedPaths.ValidateJSON)
		})

		spec.Proxy.ListenPath = "/"
	})
}

func TestValidateJSONSchema(t *testing.T) {
	ts := StartTest()
	defer ts.Close()

	testPrepareValidateJSONSchema()

	resp, _ := ts.Run(t, []test.TestCase{
		{Method: "POST", Path: "/without_validation", Data: "{not_valid}", Code: http.StatusOK},
		{Method: "POST", Path: "/v", Data: `{"age":23}`, BodyMatch: `firstName: firstName is required; lastName: lastName is required`, Code: http.StatusUnprocessableEntity},
		{Method: "POST", Path: "/v", Data: `[]`, BodyMatch: `Expected: object, given: array`, Code: http.StatusUnprocessableEntity},
		{Method: "POST", Path: "/v", Data: `not_json`, Code: http.StatusBadRequest},
		{Method: "POST", Path: "/v", Data: `{"age":23, "firstName": "Harry", "lastName": "Potter"}`, Code: http.StatusOK},
		{Method: "POST", Path: "/v", Data: `{"age":23, "firstName": "Harry", "lastName": "Potter", "objs": "d"}`, Code: http.StatusUnprocessableEntity, BodyMatch: `objs: objs must be one of the following: \\"a\\", \\"b\\", \\"c\\"`},
	}...)
	defer resp.Body.Close()
}

func BenchmarkValidateJSONSchema(b *testing.B) {
	b.ReportAllocs()

	ts := StartTest()
	defer ts.Close()

	testPrepareValidateJSONSchema()

	for i := 0; i < b.N; i++ {
		resp, _ := ts.Run(b, []test.TestCase{
			{Method: "POST", Path: "/without_validation", Data: "{not_valid}", Code: http.StatusOK},
			{Method: "POST", Path: "/v", Data: `{"age":23}`, BodyMatch: `firstName: firstName is required; lastName: lastName is required`, Code: http.StatusUnprocessableEntity},
			{Method: "POST", Path: "/v", Data: `[]`, BodyMatch: `Expected: object, given: array`, Code: http.StatusUnprocessableEntity},
			{Method: "POST", Path: "/v", Data: `not_json`, Code: http.StatusBadRequest},
			{Method: "POST", Path: "/v", Data: `{"age":23, "firstName": "Harry", "lastName": "Potter"}`, Code: http.StatusOK},
		}...)
		defer resp.Body.Close()
	}
}
