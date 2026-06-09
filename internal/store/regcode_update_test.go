package store

import "testing"

// TestUpsertRegCodeDisableExisting 锁定管理员「停用」语义所依赖的不变量：
// UpsertRegCode 的「强制 active=true」兜底只对新建（!exists）的未使用码生效，
// 对已存在的码做更新时不得把 active=false 改回 true，否则停用功能会失效。
func TestUpsertRegCodeDisableExisting(t *testing.T) {
	st := newJSONStoreForTest(t)
	if err := st.UpsertRegCode(RegCode{Code: "C1", Type: 1, Days: 30, ValidityTime: -1, UseCountLimit: 1, Active: true}); err != nil {
		t.Fatal(err)
	}
	reg, ok := st.RegCode("C1")
	if !ok {
		t.Fatal("code should exist")
	}
	reg.Active = false
	if err := st.UpsertRegCode(reg); err != nil {
		t.Fatal(err)
	}
	got, _ := st.RegCode("C1")
	if got.Active {
		t.Fatal("existing unused regcode must stay disabled after update")
	}

	// 反向：编辑有效期与次数上限应被持久化。
	got.ValidityTime = 72
	got.UseCountLimit = 5
	got.Active = true
	if err := st.UpsertRegCode(got); err != nil {
		t.Fatal(err)
	}
	after, _ := st.RegCode("C1")
	if after.ValidityTime != 72 || after.UseCountLimit != 5 || !after.Active {
		t.Fatalf("edits not persisted: %+v", after)
	}
}

// TestUpsertRegCodeNewForcesActive 保留新建未使用码被强制启用的既有兜底语义。
func TestUpsertRegCodeNewForcesActive(t *testing.T) {
	st := newJSONStoreForTest(t)
	if err := st.UpsertRegCode(RegCode{Code: "NEW", Type: 1, Days: 30, ValidityTime: -1, UseCountLimit: 1, Active: false}); err != nil {
		t.Fatal(err)
	}
	got, _ := st.RegCode("NEW")
	if !got.Active {
		t.Fatal("new unused regcode should be force-enabled on create")
	}
}
