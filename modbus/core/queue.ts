import ModbusRTU from "modbus-serial";
import { ModbusRepeater } from "./repeater";
import { ModbusAction } from "../misc";

export class ModbusQueue {
    private client: ModbusRTU;
    private actionQueue: (() => Promise<void>)[];
    private initialized = false;
    private processingActions: boolean = false;
    private disposed: boolean = false;
    constructor() {
        this.client = new ModbusRTU();
        this.actionQueue = new Array();
    }

    private async executeActions() {
        if (this.processingActions || this.disposed) {
            return;
        }
        await this.initialize();
        while (this.actionQueue.length > 0) {
            if (this.disposed) {
                return;
            }
            this.processingActions = true;
            const action = this.actionQueue.shift();
            if (!action)
                break;
            await action();
            await new Promise(r => setTimeout(r, 10));
        }
        this.processingActions = false;
    }
    async dispose() {
        ModbusRepeater.dispose();
        this.disposed = true;
        await new Promise((r) => this.client.close(r));
    }
    async initialize() {
        if (this.initialized || this.disposed) {
            return;
        }
        await this.client.connectRTUBuffered('/dev/ttyACM0', { baudRate: 9600, dataBits: 8, stopBits: 2 });
        await this.client.setTimeout(500);
        this.initialized = true;
    }
    enqueue<T = void>(action: ModbusAction<T>): Promise<T> {
        return new Promise((rs, rj) => {
            this.actionQueue.push(async (): Promise<void> => {
                await action(this.client).then(rs).catch(rj);
            });
            this.executeActions();
        });
    }
}