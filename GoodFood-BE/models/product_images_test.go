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

func testProductImages(t *testing.T) {
	t.Parallel()

	query := ProductImages()

	if query.Query == nil {
		t.Error("expected a query, got nothing")
	}
}

func testProductImagesDelete(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
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

	count, err := ProductImages().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testProductImagesQueryDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if rowsAff, err := ProductImages().DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := ProductImages().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testProductImagesSliceDeleteAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := ProductImageSlice{o}

	if rowsAff, err := slice.DeleteAll(ctx, tx); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only have deleted one row, but affected:", rowsAff)
	}

	count, err := ProductImages().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Error("want zero records, got:", count)
	}
}

func testProductImagesExists(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	e, err := ProductImageExists(ctx, tx, o.ProdutImageID)
	if err != nil {
		t.Errorf("Unable to check if ProductImage exists: %s", err)
	}
	if !e {
		t.Errorf("Expected ProductImageExists to return true, but got false.")
	}
}

func testProductImagesFind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	productImageFound, err := FindProductImage(ctx, tx, o.ProdutImageID)
	if err != nil {
		t.Error(err)
	}

	if productImageFound == nil {
		t.Error("want a record, got nil")
	}
}

func testProductImagesBind(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if err = ProductImages().Bind(ctx, tx, o); err != nil {
		t.Error(err)
	}
}

func testProductImagesOne(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	if x, err := ProductImages().One(ctx, tx); err != nil {
		t.Error(err)
	} else if x == nil {
		t.Error("expected to get a non nil record")
	}
}

func testProductImagesAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	productImageOne := &ProductImage{}
	productImageTwo := &ProductImage{}
	if err = randomize.Struct(seed, productImageOne, productImageDBTypes, false, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}
	if err = randomize.Struct(seed, productImageTwo, productImageDBTypes, false, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = productImageOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = productImageTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := ProductImages().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 2 {
		t.Error("want 2 records, got:", len(slice))
	}
}

func testProductImagesCount(t *testing.T) {
	t.Parallel()

	var err error
	seed := randomize.NewSeed()
	productImageOne := &ProductImage{}
	productImageTwo := &ProductImage{}
	if err = randomize.Struct(seed, productImageOne, productImageDBTypes, false, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}
	if err = randomize.Struct(seed, productImageTwo, productImageDBTypes, false, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = productImageOne.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}
	if err = productImageTwo.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := ProductImages().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 2 {
		t.Error("want 2 records, got:", count)
	}
}

func productImageBeforeInsertHook(ctx context.Context, e boil.ContextExecutor, o *ProductImage) error {
	*o = ProductImage{}
	return nil
}

func productImageAfterInsertHook(ctx context.Context, e boil.ContextExecutor, o *ProductImage) error {
	*o = ProductImage{}
	return nil
}

func productImageAfterSelectHook(ctx context.Context, e boil.ContextExecutor, o *ProductImage) error {
	*o = ProductImage{}
	return nil
}

func productImageBeforeUpdateHook(ctx context.Context, e boil.ContextExecutor, o *ProductImage) error {
	*o = ProductImage{}
	return nil
}

func productImageAfterUpdateHook(ctx context.Context, e boil.ContextExecutor, o *ProductImage) error {
	*o = ProductImage{}
	return nil
}

func productImageBeforeDeleteHook(ctx context.Context, e boil.ContextExecutor, o *ProductImage) error {
	*o = ProductImage{}
	return nil
}

func productImageAfterDeleteHook(ctx context.Context, e boil.ContextExecutor, o *ProductImage) error {
	*o = ProductImage{}
	return nil
}

func productImageBeforeUpsertHook(ctx context.Context, e boil.ContextExecutor, o *ProductImage) error {
	*o = ProductImage{}
	return nil
}

func productImageAfterUpsertHook(ctx context.Context, e boil.ContextExecutor, o *ProductImage) error {
	*o = ProductImage{}
	return nil
}

func testProductImagesHooks(t *testing.T) {
	t.Parallel()

	var err error

	ctx := context.Background()
	empty := &ProductImage{}
	o := &ProductImage{}

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, o, productImageDBTypes, false); err != nil {
		t.Errorf("Unable to randomize ProductImage object: %s", err)
	}

	AddProductImageHook(boil.BeforeInsertHook, productImageBeforeInsertHook)
	if err = o.doBeforeInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeInsertHook function to empty object, but got: %#v", o)
	}
	productImageBeforeInsertHooks = []ProductImageHook{}

	AddProductImageHook(boil.AfterInsertHook, productImageAfterInsertHook)
	if err = o.doAfterInsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterInsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterInsertHook function to empty object, but got: %#v", o)
	}
	productImageAfterInsertHooks = []ProductImageHook{}

	AddProductImageHook(boil.AfterSelectHook, productImageAfterSelectHook)
	if err = o.doAfterSelectHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterSelectHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterSelectHook function to empty object, but got: %#v", o)
	}
	productImageAfterSelectHooks = []ProductImageHook{}

	AddProductImageHook(boil.BeforeUpdateHook, productImageBeforeUpdateHook)
	if err = o.doBeforeUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpdateHook function to empty object, but got: %#v", o)
	}
	productImageBeforeUpdateHooks = []ProductImageHook{}

	AddProductImageHook(boil.AfterUpdateHook, productImageAfterUpdateHook)
	if err = o.doAfterUpdateHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpdateHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpdateHook function to empty object, but got: %#v", o)
	}
	productImageAfterUpdateHooks = []ProductImageHook{}

	AddProductImageHook(boil.BeforeDeleteHook, productImageBeforeDeleteHook)
	if err = o.doBeforeDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeDeleteHook function to empty object, but got: %#v", o)
	}
	productImageBeforeDeleteHooks = []ProductImageHook{}

	AddProductImageHook(boil.AfterDeleteHook, productImageAfterDeleteHook)
	if err = o.doAfterDeleteHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterDeleteHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterDeleteHook function to empty object, but got: %#v", o)
	}
	productImageAfterDeleteHooks = []ProductImageHook{}

	AddProductImageHook(boil.BeforeUpsertHook, productImageBeforeUpsertHook)
	if err = o.doBeforeUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doBeforeUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected BeforeUpsertHook function to empty object, but got: %#v", o)
	}
	productImageBeforeUpsertHooks = []ProductImageHook{}

	AddProductImageHook(boil.AfterUpsertHook, productImageAfterUpsertHook)
	if err = o.doAfterUpsertHooks(ctx, nil); err != nil {
		t.Errorf("Unable to execute doAfterUpsertHooks: %s", err)
	}
	if !reflect.DeepEqual(o, empty) {
		t.Errorf("Expected AfterUpsertHook function to empty object, but got: %#v", o)
	}
	productImageAfterUpsertHooks = []ProductImageHook{}
}

func testProductImagesInsert(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := ProductImages().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testProductImagesInsertWhitelist(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Whitelist(productImageColumnsWithoutDefault...)); err != nil {
		t.Error(err)
	}

	count, err := ProductImages().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}
}

func testProductImageToOneProductUsingProductIDProduct(t *testing.T) {
	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var local ProductImage
	var foreign Product

	seed := randomize.NewSeed()
	if err := randomize.Struct(seed, &local, productImageDBTypes, false, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}
	if err := randomize.Struct(seed, &foreign, productDBTypes, false, productColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize Product struct: %s", err)
	}

	if err := foreign.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	local.ProductID = foreign.ProductID
	if err := local.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	check, err := local.ProductIDProduct().One(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}

	if check.ProductID != foreign.ProductID {
		t.Errorf("want: %v, got %v", foreign.ProductID, check.ProductID)
	}

	ranAfterSelectHook := false
	AddProductHook(boil.AfterSelectHook, func(ctx context.Context, e boil.ContextExecutor, o *Product) error {
		ranAfterSelectHook = true
		return nil
	})

	slice := ProductImageSlice{&local}
	if err = local.L.LoadProductIDProduct(ctx, tx, false, (*[]*ProductImage)(&slice), nil); err != nil {
		t.Fatal(err)
	}
	if local.R.ProductIDProduct == nil {
		t.Error("struct should have been eager loaded")
	}

	local.R.ProductIDProduct = nil
	if err = local.L.LoadProductIDProduct(ctx, tx, true, &local, nil); err != nil {
		t.Fatal(err)
	}
	if local.R.ProductIDProduct == nil {
		t.Error("struct should have been eager loaded")
	}

	if !ranAfterSelectHook {
		t.Error("failed to run AfterSelect hook for relationship")
	}
}

func testProductImageToOneSetOpProductUsingProductIDProduct(t *testing.T) {
	var err error

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()

	var a ProductImage
	var b, c Product

	seed := randomize.NewSeed()
	if err = randomize.Struct(seed, &a, productImageDBTypes, false, strmangle.SetComplement(productImagePrimaryKeyColumns, productImageColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &b, productDBTypes, false, strmangle.SetComplement(productPrimaryKeyColumns, productColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}
	if err = randomize.Struct(seed, &c, productDBTypes, false, strmangle.SetComplement(productPrimaryKeyColumns, productColumnsWithoutDefault)...); err != nil {
		t.Fatal(err)
	}

	if err := a.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}
	if err = b.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Fatal(err)
	}

	for i, x := range []*Product{&b, &c} {
		err = a.SetProductIDProduct(ctx, tx, i != 0, x)
		if err != nil {
			t.Fatal(err)
		}

		if a.R.ProductIDProduct != x {
			t.Error("relationship struct not set to correct value")
		}

		if x.R.ProductIDProductImages[0] != &a {
			t.Error("failed to append to foreign relationship struct")
		}
		if a.ProductID != x.ProductID {
			t.Error("foreign key was wrong value", a.ProductID)
		}

		zero := reflect.Zero(reflect.TypeOf(a.ProductID))
		reflect.Indirect(reflect.ValueOf(&a.ProductID)).Set(zero)

		if err = a.Reload(ctx, tx); err != nil {
			t.Fatal("failed to reload", err)
		}

		if a.ProductID != x.ProductID {
			t.Error("foreign key was wrong value", a.ProductID, x.ProductID)
		}
	}
}

func testProductImagesReload(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
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

func testProductImagesReloadAll(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice := ProductImageSlice{o}

	if err = slice.ReloadAll(ctx, tx); err != nil {
		t.Error(err)
	}
}

func testProductImagesSelect(t *testing.T) {
	t.Parallel()

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	slice, err := ProductImages().All(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if len(slice) != 1 {
		t.Error("want one record, got:", len(slice))
	}
}

var (
	productImageDBTypes = map[string]string{`ProdutImageID`: `integer`, `Image`: `character varying`, `ProductID`: `integer`}
	_                   = bytes.MinRead
)

func testProductImagesUpdate(t *testing.T) {
	t.Parallel()

	if 0 == len(productImagePrimaryKeyColumns) {
		t.Skip("Skipping table with no primary key columns")
	}
	if len(productImageAllColumns) == len(productImagePrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := ProductImages().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImagePrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	if rowsAff, err := o.Update(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("should only affect one row but affected", rowsAff)
	}
}

func testProductImagesSliceUpdateAll(t *testing.T) {
	t.Parallel()

	if len(productImageAllColumns) == len(productImagePrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	o := &ProductImage{}
	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImageColumnsWithDefault...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Insert(ctx, tx, boil.Infer()); err != nil {
		t.Error(err)
	}

	count, err := ProductImages().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if count != 1 {
		t.Error("want one record, got:", count)
	}

	if err = randomize.Struct(seed, o, productImageDBTypes, true, productImagePrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	// Remove Primary keys and unique columns from what we plan to update
	var fields []string
	if strmangle.StringSliceMatch(productImageAllColumns, productImagePrimaryKeyColumns) {
		fields = productImageAllColumns
	} else {
		fields = strmangle.SetComplement(
			productImageAllColumns,
			productImagePrimaryKeyColumns,
		)
		fields = strmangle.SetComplement(fields, productImageGeneratedColumns)
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

	slice := ProductImageSlice{o}
	if rowsAff, err := slice.UpdateAll(ctx, tx, updateMap); err != nil {
		t.Error(err)
	} else if rowsAff != 1 {
		t.Error("wanted one record updated but got", rowsAff)
	}
}

func testProductImagesUpsert(t *testing.T) {
	t.Parallel()

	if len(productImageAllColumns) == len(productImagePrimaryKeyColumns) {
		t.Skip("Skipping table with only primary key columns")
	}

	seed := randomize.NewSeed()
	var err error
	// Attempt the INSERT side of an UPSERT
	o := ProductImage{}
	if err = randomize.Struct(seed, &o, productImageDBTypes, true); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	ctx := context.Background()
	tx := MustTx(boil.BeginTx(ctx, nil))
	defer func() { _ = tx.Rollback() }()
	if err = o.Upsert(ctx, tx, false, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert ProductImage: %s", err)
	}

	count, err := ProductImages().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}

	// Attempt the UPDATE side of an UPSERT
	if err = randomize.Struct(seed, &o, productImageDBTypes, false, productImagePrimaryKeyColumns...); err != nil {
		t.Errorf("Unable to randomize ProductImage struct: %s", err)
	}

	if err = o.Upsert(ctx, tx, true, nil, boil.Infer(), boil.Infer()); err != nil {
		t.Errorf("Unable to upsert ProductImage: %s", err)
	}

	count, err = ProductImages().Count(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	if count != 1 {
		t.Error("want one record, got:", count)
	}
}
