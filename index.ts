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
    // await relay.setMultiple({
    //     o1: true,
    //     o2: true,
    //     o3: true,
    // });
    // await new Promise((r)=>setTimeout(r, 1000));


    // await relay.setMultiple({
    //     o1: false,
    //     o2: false,
    //     o3: false,
    // });

    // await relay.off(1);
    // await queue.dispose();

    // const r1 = await client.connectRTUBuffered('/dev/ttyACM0', { baudRate: 9600, dataBits: 8, stopBits: 2 });
    // const r2 = await client.setTimeout(500);
    // const r3 = await client.setID(243);
    // const result = await client.readHoldingRegisters(128, 1);
    // console.log(result);
    // const result2 = await client.readInputRegisters(0x00C8, 6);
    // console.log(result2.buffer.toString());

    // await client.writeCoil(3, 1);

    // const coilStatus = await client.readCoils(0, 6);
    // console.log(coilStatus);

    // await new Promise((r)=>setTimeout(r, 1000));

    // await client.writeCoil(3, 0);
    // await client.close();
}

main();