import { v4 as uuidv4 } from 'uuid';

export class ModbusRepeater {
    private items: Record<string, ()=>Promise<void>>; 
    private interval: NodeJS.Timeout;
    private executing: boolean = false;
    private static instance = new ModbusRepeater();
    constructor() {
        this.items = {};
        this.interval = setInterval(()=>this.execute(), 300);
    }
    async execute() {
        if(this.executing) {
            return;
        }
        this.executing = true;
        try {
            
        } finally {
            this.executing = false;
        }
        for(const item of Object.values(this.items)) {
            await item();
        }
    }
    static enqueue(action: ()=>Promise<void>):string {
        const key = uuidv4();
        this.instance.items[key] = action;
        return key;
    }
    static dequeue(id: string){
        delete this.instance.items[id];
    }
    static dispose() {
        this.instance.items = {};
        clearInterval(this.instance.interval);
    }
}