package dao

import (
	"context"
	"errors"

	"github.com/UnderTreeTech/waterdrop/pkg/database/mysql"

	"github.com/Masterminds/squirrel"
)

type txKey struct{}

func (d *dao) Begin(ctx context.Context) (context.Context, error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return ctx, err
	}

	ctx = context.WithValue(ctx, txKey{}, tx)
	return ctx, err
}

func (d *dao) Commit(ctx context.Context) error {
	tx, err := d.GetTxFromCtx(ctx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *dao) Rollback(ctx context.Context) error {
	tx, err := d.GetTxFromCtx(ctx)
	if err != nil {
		return err
	}

	return tx.Rollback()
}

func (d *dao) GetTxFromCtx(ctx context.Context) (*mysql.Tx, error) {
	tx, ok := ctx.Value(txKey{}).(*mysql.Tx)
	if !ok {
		return nil, errors.New("assert tx err")
	}

	return tx, nil
}

func (d *dao) Analytic(build squirrel.SelectBuilder, condition map[string]interface{}) (squirrel.SelectBuilder, error) {
	// add order by
	if orderBy, ok := condition["_orderBy"]; ok {
		if orderBy, ok := orderBy.(string); ok {
			build = build.OrderBy(orderBy)
			delete(condition, "_orderBy")
		} else {
			return build, errors.New("_orderBy type is string")
		}
	}

	// add group by
	if groupBy, ok := condition["_groupBy"]; ok {
		if groupBy, ok := groupBy.(string); ok {
			build = build.GroupBy(groupBy)
			delete(condition, "_groupBy")
		} else {
			return build, errors.New("_groupBy type is string")
		}

		//add having condition
		if having, ok := condition["_having"]; ok {
			if having, ok := having.(string); ok {
				build = build.Having(having)
				delete(condition, "_having")
			} else {
				return build, errors.New("_having type is string")
			}

		}
	}

	// add offset
	if offset, ok := condition["_offset"]; ok {
		if offset, ok := offset.(uint64); ok {
			build = build.Offset(offset)
			delete(condition, "_offset")
		} else {
			return build, errors.New("_offset type is uint64")
		}

	}

	// add limit
	if limit, ok := condition["_limit"]; ok {
		if limit, ok := limit.(uint64); ok {
			build = build.Limit(limit)
			delete(condition, "_limit")
		} else {
			return build, errors.New("_limit type is uint64")
		}

	}

	return build.Where(condition), nil
}
