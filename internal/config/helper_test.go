package config

import (
	"testing"
)

func TestSetDefaultPathRewriteRules(t *testing.T) {
	config := newPathGetterTestConfig("/")
	config.setDefaultPathRewriteRules()

	if len(config.PathRewriteRules) != 4 {
		t.Fatalf("expected 4 path rewrite rules, got %d", len(config.PathRewriteRules))
	}

	if config.PathRewriteRules[0].Source != "usr/local/" ||
		config.PathRewriteRules[0].Target != "" ||
		config.PathRewriteRules[0].Operation != "REPLACE" {
		t.Errorf("rule 0 is invalid: %+v", config.PathRewriteRules[0])
	}
	if config.PathRewriteRules[1].Source != "var/lib/uniget/" ||
		config.PathRewriteRules[1].Target != "/var/lib/uniget/" ||
		config.PathRewriteRules[1].Operation != "REPLACE" ||
		config.PathRewriteRules[1].Abort != true {
		t.Errorf("rule 1 is invalid: %+v", config.PathRewriteRules[1])
	}
	if config.PathRewriteRules[2].Source != "var/cache/uniget/" ||
		config.PathRewriteRules[2].Target != "/var/cache/uniget/" ||
		config.PathRewriteRules[2].Operation != "REPLACE" ||
		config.PathRewriteRules[2].Abort != true {
		t.Errorf("rule 2 is invalid: %+v", config.PathRewriteRules[2])
	}
	if config.PathRewriteRules[3].Source != "" ||
		config.PathRewriteRules[3].Target != "usr/local/" ||
		config.PathRewriteRules[3].Operation != "PREPEND" {
		t.Errorf("rule 3 is invalid: %+v", config.PathRewriteRules[3])
	}
}

func TestSetDefaultPathRewriteRulesWithoutTarget(t *testing.T) {
	config := newPathGetterTestConfig("/")
	config.Target = ""
	config.setDefaultPathRewriteRules()

	if len(config.PathRewriteRules) != 3 {
		t.Fatalf("expected 3 path rewrite rules, got %d", len(config.PathRewriteRules))
	}
	for _, rule := range config.PathRewriteRules {
		if rule.Operation == "PREPEND" {
			t.Errorf("did not expect a PREPEND rule")
		}
	}
}

func TestSetDefaultPathRewriteRulesWithTrailingSlashTarget(t *testing.T) {
	config := newPathGetterTestConfig("/")
	config.Target = "usr/local/"
	config.setDefaultPathRewriteRules()

	if len(config.PathRewriteRules) != 4 {
		t.Fatalf("expected 4 path rewrite rules, got %d", len(config.PathRewriteRules))
	}
	if config.PathRewriteRules[3].Target != "usr/local/" {
		t.Errorf("expected PREPEND target usr/local/, got %s", config.PathRewriteRules[3].Target)
	}
}
