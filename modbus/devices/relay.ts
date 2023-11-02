import { Diff, buildDiff } from "../../utils/diff";
import { EventSender, EventSubscriber } from "../../utils/event";
import { ModbusAction } from "../misc";
import { ModbusQueue } from "../core/queue";
import { ModbusRepeater } from "../core/repeater";

interface RelayOutputs {
    o1: boolean,
    o2: boolean,
    o3: boolean,
    o4: boolean,
    o5: boolean,
    o6: boolean,
};
interface RelayInputs {
    i0: boolean,
    i1: boolean,
    i2: boolean,
    i3: boolean,
    i4: boolean,
    i5: boolean,
    i6: boolean,
};
interface RelayCounters {
    c0: number,
    c1: number,
    c2: number,
    c3: number,
    c4: number,
    c5: number,
    c6: number,
};
interface RelayState {
    out: RelayOutputs;
    in: RelayInputs;
    cnt: RelayCounters
};

const REGISTERS = {
    OUT: {
        r1: 0,
        r2: 1,
        r3: 2,
        r4: 3,
        r5: 4,
        r6: 5,
    },
    IN: {
        r0: 7,
        r1: 0,
        r2: 1,
        r3: 2,
        r4: 3,
        r5: 4,
        r6: 5,
    },
}

//https://wirenboard.com/wiki/Relay_Module_Modbus_Management
export class ModbusRelay {
    private queue: ModbusQueue;
    private id: number;
    private initialized: boolean = false;
    private state: RelayState;
    private stateChangedSender: EventSender<Diff<RelayState>>;
    constructor(queue: ModbusQueue, id: number) {
        this.stateChangedSender = new EventSender<Diff<RelayState>>();
        this.queue = queue;
        this.id = id;
        this.state = {
            cnt: { c0: 0, c1: 0, c2: 0, c3: 0, c4: 0, c5: 0, c6: 0 },
            in: { i0: false, i1: false, i2: false, i3: false, i4: false, i5: false, i6: false },
            out: { o1: false, o2: false, o3: false, o4: false, o5: false, o6: false }
        }
    }
    get stateChanged(): EventSubscriber<Diff<RelayState>> {
        return this.stateChangedSender.event;
    }
    private enqueue<T>(action: ModbusAction<T>): Promise<T> {
        return this.queue.enqueue(async cl => {
            await cl.setID(this.id);
            return await action(cl);
        })
    }
    async refreshState() {
        const state = await this.enqueue(async cl => {
            const [o1, o2, o3, o4, o5, o6] = await cl.readCoils(0, 6).then(x => x.data.slice(0, 6));
            const [i1, i2, i3, i4, i5, i6, _i_u_0, i0] = await cl.readDiscreteInputs(0, 7).then(x => x.data.slice(0, 7));
            const [c1, c2, c3, c4, c5, c6, _c_u_0, c0] = await cl.readInputRegisters(32, 8).then(x => x.data.slice(0, 8));
            return {
                out: { o1, o2, o3, o4, o5, o6 },
                in: { i1, i2, i3, i4, i5, i6, i0 },
                cnt: { c1, c2, c3, c4, c5, c6, c0 },
            };
        });
        const oldState = this.state;
        this.state = state;
        if (JSON.stringify(oldState) !== JSON.stringify(state)) {
            this.stateChangedSender.raise(buildDiff(oldState, state));
            console.debug(new Date(), this.id, 'state updated');
        }

    }
    async initialize() {
        if (this.initialized) {
            return;
        }
        this.initialized = true;
        await this.enqueue(async cl => {
            await cl.writeRegisters(9, [0, 0, 0, 0, 0, 0]);
            await cl.writeRegister(16, 3);
            console.log(new Date(), this.id, 'update config');
        });
        ModbusRepeater.enqueue(() => this.refreshState());
    }
    async on(index: number) {
        await this.enqueue(cl => cl.writeCoil(index, true));
        console.log(new Date(), this.id, 'set single (on)', index);
    }
    async off(index: number) {
        await this.enqueue(cl => cl.writeCoil(index, false));

        console.log(new Date(), this.id, 'set single (off)', index);
    }
    async setMultiple(out: Partial<RelayOutputs>) {
        const actual = { ...this.state.out, ...out };
        await this.enqueue(cl => cl.writeCoils(REGISTERS.OUT.r1, [actual.o1, actual.o2, actual.o3, actual.o4, actual.o5, actual.o6]));
        console.log(new Date(), this.id, 'set multiple');
    }
}