import { v4 as uuidv4 } from 'uuid';

export type EventHandler<TArgs> = (arg: TArgs) => Promise<void>;
export class EventSender<TArgs> {
    private pEvent: EventSubscriber<TArgs> | undefined;
    private items: { key: string, handler: EventHandler<TArgs> }[];
    constructor() {
        this.items = [];
    }
    add(handler: (arg: TArgs) => Promise<void>): string {
        const key = uuidv4();
        this.items.push({ key, handler });
        return key;
    }
    remove(key: string): void {
        const index = this.items.findIndex(x => x.key === key);
        this.items.splice(index, 1);
    }
    async raise(args: TArgs) {
        for (const { handler } of [...this.items]) {
            await handler(args);
        }
    }
    get event(): EventSubscriber<TArgs> {
        return this.pEvent ?? (this.pEvent = new EventSubscriber<TArgs>(this));
    }
}
export class EventSubscriber<TArgs> {
    private sender;
    constructor(sender: EventSender<TArgs>) {
        this.sender = sender;
    }
    add(handler: EventHandler<TArgs>): string {
        return this.sender.add(handler);
    }
    remove(key: string): void {
        return this.sender.remove(key);
    }
}