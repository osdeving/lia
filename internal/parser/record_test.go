package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile_RecordDeclaration(t *testing.T) {
	content := `module orders.core as domain {
  record OrderDTO {
    id: String;
    status: String;
    total: Decimal;
  }
}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lia")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	prog, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	if len(prog.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(prog.Modules))
	}
	mod := prog.Modules[0]
	if len(mod.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(mod.Records))
	}
	rec := mod.Records[0]
	if rec.Name != "OrderDTO" {
		t.Errorf("expected record name OrderDTO, got %s", rec.Name)
	}
	if len(rec.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(rec.Fields))
	}
	if rec.Fields[0].Name != "id" || rec.Fields[0].Type != "String" {
		t.Errorf("unexpected field 0: %s: %s", rec.Fields[0].Name, rec.Fields[0].Type)
	}
	if rec.Fields[1].Name != "status" || rec.Fields[1].Type != "String" {
		t.Errorf("unexpected field 1: %s: %s", rec.Fields[1].Name, rec.Fields[1].Type)
	}
	if rec.Fields[2].Name != "total" || rec.Fields[2].Type != "Decimal" {
		t.Errorf("unexpected field 2: %s: %s", rec.Fields[2].Name, rec.Fields[2].Type)
	}
}

func TestParseFile_RecordWithTypeAndEnum(t *testing.T) {
	content := `module data.core as domain {
  type UserId = String where nonEmpty;
  enum Status { ACTIVE, INACTIVE };
  record UserProfile {
    id: UserId;
    name: String;
    status: Status;
  }
}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lia")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	prog, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	mod := prog.Modules[0]
	if len(mod.Types) != 1 {
		t.Errorf("expected 1 type, got %d", len(mod.Types))
	}
	if len(mod.Enums) != 1 {
		t.Errorf("expected 1 enum, got %d", len(mod.Enums))
	}
	if len(mod.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(mod.Records))
	}
	if mod.Records[0].Name != "UserProfile" {
		t.Errorf("expected record name UserProfile, got %s", mod.Records[0].Name)
	}
	if len(mod.Records[0].Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(mod.Records[0].Fields))
	}
}

func TestParseFile_MultipleRecords(t *testing.T) {
	content := `module multi.rec as domain {
  record A {
    x: Int;
  }
  record B {
    y: String;
    z: Bool;
  }
}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lia")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	prog, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	mod := prog.Modules[0]
	if len(mod.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(mod.Records))
	}
	if mod.Records[0].Name != "A" || len(mod.Records[0].Fields) != 1 {
		t.Errorf("unexpected record A: %s (%d fields)", mod.Records[0].Name, len(mod.Records[0].Fields))
	}
	if mod.Records[1].Name != "B" || len(mod.Records[1].Fields) != 2 {
		t.Errorf("unexpected record B: %s (%d fields)", mod.Records[1].Name, len(mod.Records[1].Fields))
	}
}

func TestParseFile_EmptyRecord(t *testing.T) {
	content := `module empty.rec as domain {
  record Empty {
  }
}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lia")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	prog, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	mod := prog.Modules[0]
	if len(mod.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(mod.Records))
	}
	if mod.Records[0].Name != "Empty" {
		t.Errorf("expected record name Empty, got %s", mod.Records[0].Name)
	}
	if len(mod.Records[0].Fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(mod.Records[0].Fields))
	}
}

func TestParseFile_RecordGenericTypes(t *testing.T) {
	content := `module gen.rec as domain {
  record Container {
    items: list<String>;
    mapping: map<String, Int>;
  }
}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lia")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	prog, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	mod := prog.Modules[0]
	if len(mod.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(mod.Records))
	}
	rec := mod.Records[0]
	if rec.Fields[0].Type != "list<String>" {
		t.Errorf("expected list<String>, got %s", rec.Fields[0].Type)
	}
	if rec.Fields[1].Type != "map<String,Int>" {
		t.Errorf("expected map<String,Int>, got %s", rec.Fields[1].Type)
	}
}

func TestRecordToIR(t *testing.T) {
	astRec := &RecordDecl{
		Name: "TestRecord",
		Fields: FieldList{
			Fields: []*Field{
				{Name: "id", Type: TypeRef{Value: "String"}},
				{Name: "count", Type: TypeRef{Value: "Int"}},
			},
		},
	}
	irRec := recordToIR(astRec)
	if irRec.Name != "TestRecord" {
		t.Errorf("expected name TestRecord, got %s", irRec.Name)
	}
	if len(irRec.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(irRec.Fields))
	}
	if irRec.Fields[0].Name != "id" || irRec.Fields[0].Type != "String" {
		t.Errorf("unexpected field 0: %s: %s", irRec.Fields[0].Name, irRec.Fields[0].Type)
	}
}
