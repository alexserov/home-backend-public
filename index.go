package main

import (
	"context"
	"serov/home-backend-public/dataaccess"
	modbusrelay "serov/home-backend-public/modbus/devices/relay"

	"go.uber.org/zap"
)

func onRelayStateChanged(sender modbusrelay.Relay, args modbusrelay.StateChangedArgs) {
	id := uint64(sender.Id())
	userId := uint64(1)

	for switchNum, switchValue := range args.New.Inputs {
		if args.Old.Inputs[switchNum] != switchValue {
			relayRecord, err := dataaccess.GetDeviceByRelaySwitchAndUser(context.Background(), zap.L(), userId, id, uint64(switchNum))
			if err != nil {
				zap.L().Error("got error when trying to access ydb", zap.Error(err))
				continue
			}
			dataaccess.SetDeviceOnByUid(context.Background(), zap.L(), relayRecord.Id, switchValue)
		}
	}
}

func initializeRelays() {
	var relay = modbusrelay.Create(243, "modbus6chrelay")
	relay.StateChanged().Add(onRelayStateChanged)
}

func main() {
	config := zap.NewProductionConfig()
	config.DisableCaller = true
	config.Level.SetLevel(zap.DebugLevel)
	logger, _ := config.Build()

	zap.ReplaceGlobals(logger)

	initializeRelays()

	switch {
		
	}
}
