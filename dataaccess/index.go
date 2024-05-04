package dataaccess

import (
	"context"
	"reflect"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/indexed"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
	yc "github.com/ydb-platform/ydb-go-yc"
	"go.uber.org/zap"
)

func getCredentials(ctx context.Context) ydb.Option {
		return yc.WithServiceAccountKeyFileCredentials("/var/home-backend/authorized_key.json")
}

func requestInDb[TResult any](
	ctx context.Context,
	logger *zap.Logger,
	executeCallback func(s table.Session, tx *table.TransactionControl, c context.Context) (table.Transaction, result.Result, error),
	callback func(res result.Result, dbctx context.Context, resultChannel chan TResult) error,
	resultChannel chan TResult) error {
	connectCtx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()

	driver, err := ydb.Open(
		connectCtx,
		"grpcs://ydb.serverless.yandexcloud.net:2135/?database=/ru-central1/b1gslbsofjqfa1mlraqa/etnd7rgspaj688fpotaq",
		getCredentials(ctx),
	)

	if err != nil {
		logger.Error("Unable to connect to db", zap.Error(err))
		return err
	}
	defer func() { driver.Close(connectCtx) }()

	rwTx := table.TxControl(
		table.BeginTx(
			table.WithSerializableReadWrite(),
		),
		table.CommitTx(),
	)

	driver.Table().Do(connectCtx,
		func(ctx context.Context, s table.Session) error {
			_, res, err := executeCallback(s, rwTx, connectCtx)

			if err != nil {
				logger.Error("Unable to execute request", zap.Error(err))
				return err
			}

			err = callback(res, connectCtx, resultChannel)

			if err != nil {
				logger.Error("Got error from db callback", zap.Error(err))
				return err
			}

			if res.Err() != nil {
				logger.Error("Error during iteration through set", zap.Error(res.Err()))
				return res.Err()
			}

			return nil
		})
	return nil
}

type YdbUnmarshaller struct {
	field reflect.Value
}

func (um *YdbUnmarshaller) UnmarshalYDB(raw types.RawValue) error {
	val := raw.Any()
	valof := reflect.ValueOf(val)
	um.field.Set(valof)
	return nil
}

func find[R any](ctx context.Context,
	logger *zap.Logger,
	resultChannel chan R,
	executeCallback func(s table.Session, tx *table.TransactionControl, c context.Context) (table.Transaction, result.Result, error),
) error {
	return requestInDb[R](
		ctx,
		logger,
		executeCallback,
		func(res result.Result, dbctx context.Context, resultChannel chan R) error {
			rType := reflect.TypeOf(*new(R))
			fieldsCount := rType.NumField()
			dbFieldNames := make([]string, fieldsCount)

			for fieldIndex := 0; fieldIndex < fieldsCount; fieldIndex++ {
				field := rType.Field(fieldIndex)
				dbFieldName := reflect.StructTag.Get(field.Tag, "sql")
				dbFieldNames[fieldIndex] = dbFieldName
			}
			for res.NextResultSet(dbctx, dbFieldNames...) {
				for res.NextRow() {
					var rowData R
					rowDataReflectIndirect := reflect.Indirect(reflect.ValueOf(&rowData))

					pointers := make([]indexed.RequiredOrOptional, fieldsCount)
					for fieldIndex := 0; fieldIndex < fieldsCount; fieldIndex++ {
						var unmarshaller types.Scanner = &YdbUnmarshaller{field: rowDataReflectIndirect.FieldByName(rType.Field(fieldIndex).Name)}
						pointers[fieldIndex] = unmarshaller
					}

					err := res.Scan(pointers...)
					if err != nil {
						return err
					}

					logger.Debug("got value", zap.Any("value", rowData))
					resultChannel <- rowData
				}
			}
			close(resultChannel)
			return nil
		},
		resultChannel,
	)
}

func ListDevicesForUser(ctx context.Context, logger *zap.Logger, uid uint64) (*[]homeDevicesDao, error) {
	resultChannel := make(chan homeDevicesDao)
	go find(
		ctx,
		logger,
		resultChannel,
		func(s table.Session, tx *table.TransactionControl, c context.Context) (table.Transaction, result.Result, error) {
			return s.Execute(c, tx,
				`
				DECLARE $uid as Uint64;
				SELECT * from home_devices where user_id = $uid;
				`,
				table.NewQueryParameters(
					table.ValueParam("$uid", types.Uint64Value(uid)),
				),
			)
		},
	)

	result := []homeDevicesDao{}
	for resultItem := range resultChannel {
		result = append(result, resultItem)
	}

	return &result, nil
}

func GetDeviceByUid(ctx context.Context, logger *zap.Logger, uid uint64) (*homeDevicesDao, error) {
	resultChannel := make(chan homeDevicesDao)
	go find(
		ctx,
		logger,
		resultChannel,
		func(s table.Session, tx *table.TransactionControl, c context.Context) (table.Transaction, result.Result, error) {
			return s.Execute(c, tx,
				`
				DECLARE $uid as Uint64;
				SELECT * from home_devices where id = $uid;
				`,
				table.NewQueryParameters(
					table.ValueParam("$uid", types.Uint64Value(uid)),
				),
			)
		},
	)

	result := <-resultChannel

	return &result, nil
}

func GetDeviceByRelaySwitchAndUser(ctx context.Context, logger *zap.Logger, userId uint64, relayId uint64, switchId uint64) (*homeDevicesDao, error) {
	resultChannel := make(chan homeDevicesDao)
	go find(
		ctx,
		logger,
		resultChannel,
		func(s table.Session, tx *table.TransactionControl, c context.Context) (table.Transaction, result.Result, error) {
			return s.Execute(c, tx,
				`
				DECLARE $userId as Uint64;
				DECLARE $relayId as Uint64;
				DECLARE $switchId as Uint64;

				SELECT * from home_devices where user_id = $uid and relay_id = $relayId and switch_id = $switchId;
				`,
				table.NewQueryParameters(
					table.ValueParam("$userId", types.Uint64Value(userId)),
					table.ValueParam("$relayId", types.Uint64Value(relayId)),
					table.ValueParam("$switchId", types.Uint64Value(switchId)),
				),
			)
		},
	)

	result := <-resultChannel

	return &result, nil
}

func SetDeviceOnByUid(ctx context.Context, logger *zap.Logger, uid uint64, value bool) (error) {
	go requestInDb[interface{}](
		ctx,
		logger,
		func(s table.Session, tx *table.TransactionControl, c context.Context) (table.Transaction, result.Result, error) {
			return s.Execute(c, tx,
				`
				DECLARE $uid as Uint64;
				DECLARE $value as Bool
				UPDATE home_devices
				SET on = value
				WHERE id = $uid;
				`,
				table.NewQueryParameters(
					table.ValueParam("$uid", types.Uint64Value(uid)),
					table.ValueParam("$value", types.BoolValue(value)),

				),
			)
		},
		func(res result.Result, dbctx context.Context, resultChannel chan interface{}) error {
			return nil
		},
		nil,
	)

	return nil
}