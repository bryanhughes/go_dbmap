package dbmap

import (
	"testing"
)

func TestReadConfig(t *testing.T) {
	var cfg Config
 	ReadFile(&cfg, "test_config.yml")

	testDatabase(cfg, t)
	testOutput(cfg, t)
	testProto(cfg, t)
	testGenerator(cfg, t)
}

func testGenerator(cfg Config, t *testing.T) {
	if len(cfg.Generator.Schemas) == 0 {
		t.Errorf("generator.schemas must specify at least one schema")
	}

	if len(cfg.Generator.ExcludedTables) != 0 {
		// Make sure the set is good
		for i, e := range cfg.Generator.ExcludedTables {
			if e == "" {
				t.Errorf("%d) generator:excluded_tables must have a value", i)
			}
		}
	} else {
		t.Error("test config does not include excluded_tables")
	}

	if len(cfg.Generator.ExcludedColumns) != 0 {
		// Make sure the set is good
		for i, e := range cfg.Generator.ExcludedColumns {
			if e.Tablename == "" {
				t.Errorf("%d) generator:excluded_columns:table must have a value", i)
			}

			if len(e.Columns) == 0 {
				t.Errorf("%d) generator:excluded_columns:columns must have at least one column", i)
			}
		}
	} else {
		t.Error("test config does not include excluded_columns")
	}

	if len(cfg.Generator.Mapping) != 0 {
		for i, e := range cfg.Generator.Mapping {
			if e.Tablename == "" {
				t.Errorf("%d) generator:mapping:table must have a value", i)
			}

			if len(e.Queries) != 0 {
				for i, e := range e.Queries {
					if e.Name == "" {
						t.Errorf("%d) generator:mapping:table:name must have a value", i)
					}

					if e.Query == "" {
						t.Errorf("%d) generator:mapping:table:name must have a value", i)
					}
				}
			}
		}
	} else {
		t.Error("test config does not include mappings")
	}

	if len(cfg.Generator.Transforms) != 0 {
		for i, e := range cfg.Generator.Transforms {
			if e.Tablename == "" {
				t.Errorf("%d) generator:xforms:table must have a value", i)
			}

			testXforms("select", e.Xforms.Select, t)
			testXforms("insert", e.Xforms.Insert, t)
			testXforms("update", e.Xforms.Update, t)
		}
	} else {
		t.Error("test config does not include xforms")
	}
}

func testXforms(which string, xforms []struct {
	Columnname string `yaml:"column"`
	Datatype   string `yaml:"data_type"`
	Xform      string `yaml:"xform"`
}, t *testing.T) {
	if len(xforms) != 0 {
		for i, ee := range xforms {
			if ee.Columnname == "" {
				t.Errorf("%d) generator:xforms:%s:column must have a value", i, which)
			}
			if ee.Datatype == "" {
				t.Errorf("%d) generator:xforms:%s:data_type must have a value", i, which)
			}
			if ee.Xform == "" {
				t.Errorf("%d) generator:xforms:%s:xform must have a value", i, which)
			}
		}
	}
}

func testProto(cfg Config, t *testing.T) {
	if cfg.Proto.Path == "" {
		t.Errorf("proto.path must have a value")
	}

	if cfg.Proto.JavaPackage == "" {
		t.Errorf("proto.java_package must have a value")
	}

	if cfg.Proto.Version == "" {
		t.Errorf("proto.version must have a value")
	}
}

func testOutput(cfg Config, t *testing.T) {
	if cfg.Output.Path == "" {
		t.Errorf("output.path must have a value")
	}

	if cfg.Output.Suffix == "" {
		t.Errorf("output.suffix must have a value")
	}

	if cfg.Output.Lang == "" {
		t.Errorf("output.lang must have a value")
	}
}

func testDatabase(cfg Config, t *testing.T) {
	if cfg.Database.Provider == "" {
		t.Errorf("database.provider must have a value")
	}

	if cfg.Database.Host == "" {
		t.Errorf("database.host must have a value")
	}

	if cfg.Database.Database == "" {
		t.Errorf("database.database must have a value")
	}

	if cfg.Database.Port == "" {
		t.Errorf("database.port must have a value")
	}

	if cfg.Database.Username == "" {
		t.Errorf("database.user must have a value")
	}

	if cfg.Database.Password == "" {
		t.Errorf("database.password must have a value")
	}
}