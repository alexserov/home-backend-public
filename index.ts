import { Diff, buildDiff } from './utils/diff'
import { ModbusAction } from './modbus/misc';
import { ModbusQueue } from './modbus/core/queue';
import { ModbusRelay } from './modbus/devices/relay';



async function main() {
    const queue = new ModbusQueue();
    const relay = new ModbusRelay(queue, 243);
    await relay.initialize();

    relay.stateChanged.add(async (diff)=>{
        if(diff.cnt.changed && diff.cnt.value.c0.changed) {
            if(diff.out.value.o1.newValue) {
                await relay.setMultiple({
                    o1: false,
                    o2: false,
                    o3: false,
                });
            } else {
                await relay.setMultiple({
                    o1: true,
                    o2: true,
                    o3: true,
                });
            }
        }
    })
}

main();