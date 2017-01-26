package tarantool

import "testing"

// These are just sanity tests

func TestValidateGood(t *testing.T) {
	schema := `{
    "type": "record",
    "name": "User",
    "fields": [
      {"name": "username", "type": "string"},
      {"name": "phone", "type": "long"},
      {"name": "age", "type": "int"}
    ]
  }`

	if err := Validate(schema); err != nil {
		t.Fatalf("expected schema to pass validation: %s", err)
	}
}

func TestValidateBad(t *testing.T) {
	schemas := []string{
		// invalid `array` type above
		`{
	    "type": "record",
	    "name": "User",
	    "fields": [
	      {"name": "username", "type": "array"},
	      {"name": "phone", "type": "long"},
	      {"name": "age", "type": "int"}
	    ]
	  }`,
		// invalid json (notice username has only one quotes)
		`{
	    "type": "record",
	    "name": "User",
	    "fields": [
	      {"name": username", "type": "string"},
	      {"name": "phone", "type": "long"},
	      {"name": "age", "type": "int"}
	    ]
	  }`,
	}

	for _, schema := range schemas {
		if err := Validate(schema); err == nil {
			t.Fatalf("expected schema to fail validation but it passed")
		}
	}
}
