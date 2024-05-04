package main

import (
	"context"
	"serov/home-backend-public/dataaccess"
	modbusrelay "serov/home-backend-public/modbus/devices/relay"
	"strconv"
	"time"

	"go.uber.org/zap"
)

var relays = map[uint64]modbusrelay.Relay{}

func onRelayStateChanged(sender modbusrelay.Relay, args modbusrelay.StateChangedArgs) {
	id := uint64(sender.Id())
	userId := uint64(1)

	zap.L().Debug("state changed", zap.Uint64("switch_id", id), zap.Any("state", args))

	for switchNum, switchValue := range args.New.Outputs {
		if args.Old.Outputs[switchNum] != switchValue {
			relayRecord, err := dataaccess.GetDeviceByRelaySwitchAndUser(context.Background(), zap.L(), userId, id, uint64(switchNum))
			if err != nil {
				zap.L().Error("got error when trying to access ydb", zap.Error(err))
				continue
			}
			dataaccess.SetDeviceOnByUid(context.Background(), zap.L(), relayRecord.Id, switchValue)
		}
	}
}

func initializeRelay(id byte) {
	var relay = modbusrelay.Create(id, "modbus6chrelay")
	relay.StateChanged().Add(onRelayStateChanged)
	relays[uint64(id)] = relay
}

func initializeRelays() {
	initializeRelay(243)
}

func fetchAndProcessCommands() {
	commands, err := dataaccess.ListCommandsForUser(context.Background(), zap.L(), 1)
	if err != nil {
		zap.L().Error("unable to fetch commands", zap.Error(err))
	}

	for _, command := range *commands {
		zap.L().Debug("got command", zap.Any("command", command))
		dataaccess.DeleteCommand(context.Background(), zap.L(), command.Id)

		device, err := dataaccess.GetDeviceByUid(context.Background(), zap.L(), command.DeviceId)

		if err != nil {
			zap.L().Error("unable to fetch device for command (got error)", zap.Error(err), zap.Any("command", command))
		}
		if device == nil {
			zap.L().Error("unable to fetch device for command", zap.Any("command", command))
			continue
		}

		relayItem := relays[device.RelayId]

		if relayItem == nil {
			zap.L().Error("no relay for device found", zap.Any("device", device))
			continue
		}

		newValue, err := strconv.ParseBool(command.Value)

		if err != nil {
			zap.L().Error("value is not bool", zap.Any("command", command))
			continue
		}

		relayItem.Set(byte(device.SwitchId), newValue)
	}
}

func listenCommands() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				fetchAndProcessCommands()
			}
		}
	}()
}

func main() {
	config := zap.NewProductionConfig()
	config.DisableCaller = true
	config.Level.SetLevel(zap.DebugLevel)
	logger, _ := config.Build()

	zap.ReplaceGlobals(logger)

	zap.L().Debug("start")

	initializeRelays()
	listenCommands()

	select {}
}
