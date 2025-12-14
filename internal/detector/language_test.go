package detector

import "testing"

func TestDetectLanguage_Go(t *testing.T) {
	lang := DetectLanguage("main.go")
	if lang != "go" {
		t.Errorf("expected 'go', got '%s'", lang)
	}
}

func TestDetectLanguage_Python(t *testing.T) {
	lang := DetectLanguage("script.py")
	if lang != "python" {
		t.Errorf("expected 'python', got '%s'", lang)
	}
}

func TestDetectLanguage_JavaScript(t *testing.T) {
	lang := DetectLanguage("app.js")
	if lang != "javascript" {
		t.Errorf("expected 'javascript', got '%s'", lang)
	}
}

func TestDetectLanguage_TypeScript(t *testing.T) {
	lang := DetectLanguage("component.ts")
	if lang != "typescript" {
		t.Errorf("expected 'typescript', got '%s'", lang)
	}

	// Also test .tsx
	langTsx := DetectLanguage("component.tsx")
	if langTsx != "typescript" {
		t.Errorf("expected 'typescript' for .tsx, got '%s'", langTsx)
	}
}

func TestDetectLanguage_Java(t *testing.T) {
	lang := DetectLanguage("Main.java")
	if lang != "java" {
		t.Errorf("expected 'java', got '%s'", lang)
	}
}

func TestDetectLanguage_Rust(t *testing.T) {
	lang := DetectLanguage("lib.rs")
	if lang != "rust" {
		t.Errorf("expected 'rust', got '%s'", lang)
	}
}

func TestDetectLanguage_Ruby(t *testing.T) {
	lang := DetectLanguage("app.rb")
	if lang != "ruby" {
		t.Errorf("expected 'ruby', got '%s'", lang)
	}
}

func TestDetectLanguage_C(t *testing.T) {
	lang := DetectLanguage("main.c")
	if lang != "c" {
		t.Errorf("expected 'c', got '%s'", lang)
	}
}

func TestDetectLanguage_Cpp(t *testing.T) {
	// Test .cpp
	lang := DetectLanguage("main.cpp")
	if lang != "cpp" {
		t.Errorf("expected 'cpp', got '%s'", lang)
	}

	// Test .cc
	langCc := DetectLanguage("main.cc")
	if langCc != "cpp" {
		t.Errorf("expected 'cpp' for .cc, got '%s'", langCc)
	}

	// Test .cxx
	langCxx := DetectLanguage("main.cxx")
	if langCxx != "cpp" {
		t.Errorf("expected 'cpp' for .cxx, got '%s'", langCxx)
	}
}

func TestDetectLanguage_Header(t *testing.T) {
	lang := DetectLanguage("header.h")
	if lang != "c" {
		t.Errorf("expected 'c', got '%s'", lang)
	}

	// Test .hpp
	langHpp := DetectLanguage("header.hpp")
	if langHpp != "cpp" {
		t.Errorf("expected 'cpp' for .hpp, got '%s'", langHpp)
	}
}

func TestDetectLanguage_Markdown(t *testing.T) {
	lang := DetectLanguage("README.md")
	if lang != "markdown" {
		t.Errorf("expected 'markdown', got '%s'", lang)
	}
}

func TestDetectLanguage_JSON(t *testing.T) {
	lang := DetectLanguage("config.json")
	if lang != "json" {
		t.Errorf("expected 'json', got '%s'", lang)
	}
}

func TestDetectLanguage_YAML(t *testing.T) {
	// Test .yaml
	lang := DetectLanguage("config.yaml")
	if lang != "yaml" {
		t.Errorf("expected 'yaml', got '%s'", lang)
	}

	// Test .yml
	langYml := DetectLanguage("config.yml")
	if langYml != "yaml" {
		t.Errorf("expected 'yaml' for .yml, got '%s'", langYml)
	}
}

func TestDetectLanguage_TOML(t *testing.T) {
	lang := DetectLanguage("config.toml")
	if lang != "toml" {
		t.Errorf("expected 'toml', got '%s'", lang)
	}
}

func TestDetectLanguage_Unknown(t *testing.T) {
	lang := DetectLanguage("file.xyz")
	if lang != "unknown" {
		t.Errorf("expected 'unknown', got '%s'", lang)
	}
}

func TestDetectLanguage_JSX(t *testing.T) {
	lang := DetectLanguage("component.jsx")
	if lang != "javascript" {
		t.Errorf("expected 'javascript' for .jsx, got '%s'", lang)
	}
}

func TestDetectLanguage_FullPath(t *testing.T) {
	// Should work with full paths
	lang := DetectLanguage("/path/to/project/src/main.go")
	if lang != "go" {
		t.Errorf("expected 'go', got '%s'", lang)
	}
}
