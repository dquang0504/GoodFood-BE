// Code generated by SQLBoiler 4.18.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/volatiletech/randomize"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/strmangle"
)

var (
	// Relationships sometimes use the reflection helper queries.Equal/queries.Assign
	// so force a package dependency in case they don't.
	_ = queries.Equal
)

func testAddresses(t *testing.T) {
	t.Parallel()

	query := Addresses()

	if query.Query == nil {
		t.Error("expected a query, got nothing")
	}
}

func testAddressesDelete(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := o.Delete(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Addresses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testAddressesQueryDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := Addresses().DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Addresses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testAddressesSliceDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := AddressSlice{o}

	if rowsAff, err := slice.DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := Addresses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testAddressesExists(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	e, err := AddressExists(ctx, tx, o.AddressID)
	if err != nil {
		t.Errorf("Unable to check if Address exists: %s", err)
	}
	if !e {
		t.Errorf("Expected AddressExists to return true, but got false.")
	}
}

func testAddressesFind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	addressFound, err := FindAddress(ctx, tx, o.AddressID)
	if err != nil {
		t.Error(err)
	}

	if addressFound == nil {
		t.Error("want a record, got nil")
	}
}

func testAddressesBind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = Addresses().Bind(ctx, tx, o); err != nil {
		t.Error(err)
	}
}

func testAddressesOne(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if x, err := Addresses().One(ctx, tx); err != nil {
		t.Error(err)
	} else if x == nil {
		t.Error("expected to get a non nil record")
	}
}

func testAddressesAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	addressOne := &Address{}
	addressTwo := &Address{}
	if err = randomize.Struct(seed, addressOne, addressDBTypes, false, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}
	if err = randomize.Struct(seed, addressTwo, addressDBTypes, false, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = addressOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = addressTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := Addresses().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 2 {
		t.Error("want 2 records, got:", len(slice))
	}
}

func testAddressesCount(t *testing.T) {
	t.Parallel()

	var err error
	seed := randomize.NewSeed()
	addressOne := &Address{}
	addressTwo := &Address{}
	if err = randomize.Struct(seed, addressOne, addressDBTypes, false, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}
	if err = randomize.Struct(seed, addressTwo, addressDBTypes, false, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = addressOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = addressTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Addresses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 2 {
		t.Error("want 2 records, got:", count)
	}
}

func addressBeforeInsertHook(ctx context.Context, e boil.ContextExecutor, o *Address) error {
	*o = Address{}
	return nil
}

func addressAfterInsertHook(ctx context.Context, e boil.ContextExecutor, o *Address) error {
	*o = Address{}
	return nil
}

func addressAfterSelectHook(ctx context.Context, e boil.ContextExecutor, o *Address) error {
	*o = Address{}
	return nil
}

func addressBeforeUpdateHook(ctx context.Context, e boil.ContextExecutor, o *Address) error {
	*o = Address{}
	return nil
}

func addressAfterUpdateHook(ctx context.Context, e boil.ContextExecutor, o *Address) error {
	*o = Address{}
	return nil
}

func addressBeforeDeleteHook(ctx context.Context, e boil.ContextExecutor, o *Address) error {
	*o = Address{}
	return nil
}

func addressAfterDeleteHook(ctx context.Context, e boil.ContextExecutor, o *Address) error {
	*o = Address{}
	return nil
}

func addressBeforeUpsertHook(ctx context.Context, e boil.ContextExecutor, o *Address) error {
	*o = Address{}
	return nil
}

func addressAfterUpsertHook(ctx context.Context, e boil.ContextExecutor, o *Address) error {
	*o = Address{}
	return nil
}

func testAddressesHooks(t *testing.T) {
	t.Parallel()

	var err error

	ctx := context.Background()
	empty := &Address{}
	o := &Address{}

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, o, addressDBTypes, false); err != nil {
		t.Errorf("Unable to randomize Address object: %s", err)
	}

	AddAddressHook(boil.BeforeInsertHook, addressBeforeInsertHook)
	if err = o.doBeforeInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeInsertHook function to empty object, but got: %#v", o)
	}
	addressBeforeInsertHooks = []AddressHook{}

	AddAddressHook(boil.AfterInsertHook, addressAfterInsertHook)
	if err = o.doAfterInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterInsertHook function to empty object, but got: %#v", o)
	}
	addressAfterInsertHooks = []AddressHook{}

	AddAddressHook(boil.AfterSelectHook, addressAfterSelectHook)
	if err = o.doAfterSelectHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterSelectHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterSelectHook function to empty object, but got: %#v", o)
	}
	addressAfterSelectHooks = []AddressHook{}

	AddAddressHook(boil.BeforeUpdateHook, addressBeforeUpdateHook)
	if err = o.doBeforeUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpdateHook function to empty object, but got: %#v", o)
	}
	addressBeforeUpdateHooks = []AddressHook{}

	AddAddressHook(boil.AfterUpdateHook, addressAfterUpdateHook)
	if err = o.doAfterUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpdateHook function to empty object, but got: %#v", o)
	}
	addressAfterUpdateHooks = []AddressHook{}

	AddAddressHook(boil.BeforeDeleteHook, addressBeforeDeleteHook)
	if err = o.doBeforeDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeDeleteHook function to empty object, but got: %#v", o)
	}
	addressBeforeDeleteHooks = []AddressHook{}

	AddAddressHook(boil.AfterDeleteHook, addressAfterDeleteHook)
	if err = o.doAfterDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterDeleteHook function to empty object, but got: %#v", o)
	}
	addressAfterDeleteHooks = []AddressHook{}

	AddAddressHook(boil.BeforeUpsertHook, addressBeforeUpsertHook)
	if err = o.doBeforeUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpsertHook function to empty object, but got: %#v", o)
	}
	addressBeforeUpsertHooks = []AddressHook{}

	AddAddressHook(boil.AfterUpsertHook, addressAfterUpsertHook)
	if err = o.doAfterUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpsertHook function to empty object, but got: %#v", o)
	}
	addressAfterUpsertHooks = []AddressHook{}
}

func testAddressesInsert(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Addresses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testAddressesInsertWhitelist(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Whitelist(addressColumnsWithoutDefault...)); err != nil {
		t.Error(err)
	}

	count, err := Addresses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testAddressToOneAccountUsingAccountIDAccount(t *testing.T) {
	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var local Address
	var foreign Account

	seed := randomize.NewSeed()
	if err := randomize.Struct(seed, &local, addressDBTypes, false, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}
	if err := randomize.Struct(seed, &foreign, accountDBTypes, false, accountColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Account struct: %s", err)
	}

	if err := foreign.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	local.AccountID = foreign.AccountID
	if err := local.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	check, err := local.AccountIDAccount().One(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}

	if check.AccountID != foreign.AccountID {
		t.Errorf("want: %v, got %v", foreign.AccountID, check.AccountID)
	}

	ranAfterSelectHook := false
	AddAccountHook(boil.AfterSelectHook, func(ctx context.Context, e boil.ContextExecutor, o *Account) error {
		ranAfterSelectHook = true
		return nil
	})

	slice := AddressSlice{&local}
	if err = local.L.LoadAccountIDAccount(ctx, tx, false, (*[]*Address)(&slice), nil); err != nil {
		t.Fatal(err)
	}
	if local.R.AccountIDAccount == nil {
		t.Error("struct should have been eager loaded")
	}

	local.R.AccountIDAccount = nil
	if err = local.L.LoadAccountIDAccount(ctx, tx, true, &local, nil); err != nil {
		t.Fatal(err)
	}
	if local.R.AccountIDAccount == nil {
		t.Error("struct should have been eager loaded")
	}

	if !ranAfterSelectHook {
		t.Error("failed to run AfterSelect hook for relationship")
	}
}

func testAddressToOneSetOpAccountUsingAccountIDAccount(t *testing.T) {
	var err error

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var a Address
	var b, c Account

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, &a, addressDBTypes, false, strmangle.SetComplement(addressPrimaryKeyColumns, addressColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &b, accountDBTypes, false, strmangle.SetComplement(accountPrimaryKeyColumns, accountColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &c, accountDBTypes, false, strmangle.SetComplement(accountPrimaryKeyColumns, accountColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}

	if err := a.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = b.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	for i, x := range []*Account{&b, &c} {
		err = a.SetAccountIDAccount(ctx, tx, i != 0, x)
		if err != nil {
			t.Fatal(err)
		}

		if a.R.AccountIDAccount != x {
			t.Error("relationship struct not set to correct value")
		}

		if x.R.AccountIDAddresses[0] != &a {
			t.Error("failed to append to foreign relationship struct")
		}
		if a.AccountID != x.AccountID {
			t.Error("foreign key was wrong value", a.AccountID)
		}

		zero := reflect.Zero(reflect.TypeOf(a.AccountID))
		reflect.Indirect(reflect.ValueOf(&a.AccountID)).Set(zero)

		if err = a.Reload(ctx, tx); err != nil {
			t.Fatal("failed to reload", err)
		}

		if a.AccountID != x.AccountID {
			t.Error("foreign key was wrong value", a.AccountID, x.AccountID)
		}
	}
}

func testAddressesReload(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = o.Reload(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testAddressesReloadAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := AddressSlice{o}

	if err = slice.ReloadAll(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testAddressesSelect(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := Addresses().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 1 {
		t.Error("want one record, got:", len(slice))
	}
}

var (
	addressDBTypes = map[string]string{`AddressID`: `integer`, `PhoneNumber`: `character varying`, `FullName`: `character varying`, `Address`: `character varying`, `SpecificAddress`: `character`, `Status`: `boolean`, `ProvinceID`: `integer`, `DistrictID`: `integer`, `WardID`: `integer`, `DeleteStatus`: `boolean`, `AccountID`: `integer`, `WardCode`: `character varying`}
	_              = bytes.MinRead
)

func testAddressesUpdate(t *testing.T) {
	t.Parallel()

	if 0 == len(addressPrimaryKeyColumns) {
		t.Skip("Skipping table with no primary key columns")
	}
	if len(addressAllColumns) == len(addressPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Addresses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, addressDBTypes, true, addressPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	if rowsAff, err := o.Update(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only affect one row but affected", rowsAff)
	}
}

func testAddressesSliceUpdateAll(t *testing.T) {
	t.Parallel()

	if len(addressAllColumns) == len(addressPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &Address{}
	if err = randomize.Struct(seed, o, addressDBTypes, true, addressColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := Addresses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, addressDBTypes, true, addressPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	// Remove Primary keys and unique columns from what we plan to update
	var fields []string
	if strmangle.StringSliceMatch(addressAllColumns, addressPrimaryKeyColumns) {
		fields = addressAllColumns
	} else {
		fields = strmangle.SetComplement(
			addressAllColumns,
			addressPrimaryKeyColumns,
		)
		fields = strmangle.SetComplement(fields, addressGeneratedColumns)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	typ := reflect.TypeOf(o).Elem()
	n := typ.NumField()

	updateMap := M{}
	for _, col := range fields {
		for i := 0; i < n; i++ {
			f := typ.Field(i)
			if f.Tag.Get("boil") == col {
				updateMap[col] = value.Field(i).Interface()
			}
		}
	}

	slice := AddressSlice{o}
	if rowsAff, err := slice.UpdateAll(ctx, tx, updateMap); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("wanted one record updated but got", rowsAff)
	}
}

func testAddressesUpsert(t *testing.T) {
	t.Parallel()

	if len(addressAllColumns) == len(addressPrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	// Attempt the INSERT side of an UPSERT
	o := Address{}
	if err = randomize.Struct(seed, &o, addressDBTypes, true); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Upsert(ctx, tx, false, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert Address: %s", err)
	}

	count, err := Addresses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}

	// Attempt the UPDATE side of an UPSERT
	if err = randomize.Struct(seed, &o, addressDBTypes, false, addressPrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize Address struct: %s", err)
	}

	if err = o.Upsert(ctx, tx, true, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert Address: %s", err)
	}

	count, err = Addresses().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}
}
