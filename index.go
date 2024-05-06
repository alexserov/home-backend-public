package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"serov/home-backend-public/dataaccess"
	modbusrelay "serov/home-backend-public/modbus/devices/relay"
	"strconv"
	"time"

	"go.uber.org/zap"
)

var relays = map[uint64]modbusrelay.Relay{}
var dialogId *string = nil
var dialogsOauthKey * string = nil

func onRelayStateChanged(sender modbusrelay.Relay, args modbusrelay.StateChangedArgs) {
	id := uint64(sender.Id())
	userId := uint64(1)

	zap.L().Debug("state changed", zap.Uint64("switch_id", id), zap.Any("state", args))

	for switchNum, switchValue := range args.New.Outputs {
		if args.Old.Outputs[switchNum] != switchValue {
			updateRelaySwitchDb(userId, id, switchNum, switchValue)
		}
	}
}

func updateRelaySwitchDb(userId uint64, relayId uint64, switchNum int, switchValue bool) {
	relayRecord, err := dataaccess.GetDeviceByRelaySwitchAndUser(context.Background(), zap.L(), userId, relayId, uint64(switchNum+1))
	if err != nil {
		zap.L().Error("got error when trying to access ydb", zap.Error(err))
		return
	}
	dataaccess.SetDeviceOnByUid(context.Background(), zap.L(), relayRecord.Id, switchValue)

	client := &http.Client{}
	payload := []byte(fmt.Sprintf(`{
		"ts": %v,
		"payload": {
			"user_id": "%v",
			"devices": [{
				"id": "%v",
				"capabilities": [
					{
					   "type": "devices.capabilities.on_off",
					   "state": {
						 "instance": "on",
						 "value": %v
					   }
					}
				  ]
			}]
		}
	}`, time.Now().Unix(), userId, relayRecord.Id, switchValue))
	req, err := http.NewRequest("POST", fmt.Sprintf("https://dialogs.yandex.net/api/v1/skills/%v/callback/state", *dialogId), bytes.NewBuffer(payload))
	if err != nil {
		zap.L().Error("updateRelaySwitchDb: unable to build request", zap.Error(err))
		panic("build-request")
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("OAuth %v", dialogsOauthKey))

	zap.L().Debug("updateRelaySwitchDb: request", zap.Any("request", req))

	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error("updateRelaySwitchDb: unable to execute request", zap.Error(err))
		panic("execute-request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error("updateRelaySwitchDb: unable to read response", zap.Error(err))
		return
	}

	zap.L().Debug("updateRelaySwitchDb: response body", zap.String("body", string(body)))
}

func initializeRelay(id byte) {
	var relay = modbusrelay.Create(id, "modbus6chrelay")
	relay.StateChanged().Add(onRelayStateChanged)
	relays[uint64(id)] = relay
	for num, val := range relay.State().Outputs {
		updateRelaySwitchDb(1, uint64(id), num, val)
	}
}

func initializeRelays() {
	initializeRelay(243)
	initializeRelay(61)
	initializeRelay(52)
	initializeRelay(53)
}

func fetchAndProcessCommands() {
	commands, err := dataaccess.ListCommandsForUser(context.Background(), zap.L(), 1)
	if err != nil {
		zap.L().Error("unable to fetch commands", zap.Error(err))
	}

	for _, command := range *commands {
		go func(cmd *dataaccess.HomeDeviceTasksDao) {
			zap.L().Debug("got command", zap.Any("command", cmd))
			dataaccess.DeleteCommand(context.Background(), zap.L(), cmd.Id)

			device, err := dataaccess.GetDeviceByUid(context.Background(), zap.L(), cmd.DeviceId)

			if err != nil {
				zap.L().Error("unable to fetch device for command (got error)", zap.Error(err), zap.Any("command", cmd))
			}
			if device == nil {
				zap.L().Error("unable to fetch device for command", zap.Any("command", cmd))
				return
			}

			relayItem := relays[device.RelayId]

			if relayItem == nil {
				zap.L().Error("no relay for device found", zap.Any("device", device))
				return
			}

			newValue, err := strconv.ParseBool(cmd.Value)

			if err != nil {
				zap.L().Error("value is not bool", zap.Any("command", cmd))
				return
			}

			zap.L().Debug("new value", zap.Any("relay id", relayItem.Id()), zap.Any("switch", device.SwitchId), zap.Any("value", newValue))
			relayItem.Set(byte(device.SwitchId-1), newValue)
		}(&command)

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

func ensureInternetConnection() {
	ok := false
	for i := 0; i < 20; i++ {
		res, err := http.Get("https://ya.ru")

		if err != nil {
			zap.L().Error("got error when trying to check internet connection", zap.Error(err))
			time.Sleep(1 * time.Second)
			continue
		}
		if res.StatusCode >= 400 {
			zap.L().Error("bad status when trying to check internet connection", zap.Any("response", res))
			time.Sleep(1 * time.Second)
			continue
		}
		ok = true
		break
	}
	if !ok {
		panic("unable to ensure internet connection")
	}
	zap.L().Debug("got internet connection!")
}

func getSecrets() {
	dialogIdLocal, err := dataaccess.GetSecret(context.Background(), zap.L(), "dialog-id")

	if err != nil {
		zap.L().Error("unable to get dialogId", zap.Error(err))
		panic("dialog-id")
	}
	dialogId = &dialogIdLocal

	dialogsOauthKeyLocal, err := dataaccess.GetSecret(context.Background(), zap.L(), "dialogs-oauth-key")

	if err != nil {
		zap.L().Error("unable to get dialogsOauthKey", zap.Error(err))
		panic("dialogs-oauth-key")
	}
	dialogsOauthKey = &dialogsOauthKeyLocal
}

func main() {
	config := zap.NewProductionConfig()
	config.DisableCaller = true
	config.Level.SetLevel(zap.DebugLevel)
	logger, _ := config.Build()

	zap.ReplaceGlobals(logger)

	zap.L().Debug("start")

	ensureInternetConnection()
	getSecrets()
	

	initializeRelays()
	listenCommands()

	select {}
}
